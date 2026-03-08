package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	"github.com/jackc/pgx/v5"
)

func (service *MerchantManager) BootstrapActor(ctx context.Context, merchantID string, actor ports.BootstrapActor) (pgsqlc.CreateActorRow, *domainerr.DomainError) {
	if service.db == nil {
		return pgsqlc.CreateActorRow{}, domainerr.NewInternalError(errors.New("merchant manager database is not configured"))
	}

	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return pgsqlc.CreateActorRow{}, parseErr
	}

	roleType := pgsqlc.RoleType(strings.ToLower(strings.TrimSpace(actor.Role)))
	switch roleType {
	case pgsqlc.RoleTypeMerchant, pgsqlc.RoleTypeEmployee, pgsqlc.RoleTypeCustomer:
	default:
		return pgsqlc.CreateActorRow{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid bootstrap role", fmt.Errorf("invalid bootstrap role"))
	}

	hashedPassword, hashErr := utils.HashPassword(actor.Password)
	if hashErr != nil {
		return pgsqlc.CreateActorRow{}, hashErr
	}

	firstName, lastName := splitFullName(actor.FullName)
	var createdActor pgsqlc.CreateActorRow

	tx, beginErr := service.db.BeginTx(ctx, pgx.TxOptions{})
	if beginErr != nil {
		return pgsqlc.CreateActorRow{}, domainerr.NewInternalError(beginErr)
	}
	store := pgsqlc.New(tx)

	rollbackOnErr := func(domainErr *domainerr.DomainError) (pgsqlc.CreateActorRow, *domainerr.DomainError) {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return pgsqlc.CreateActorRow{}, domainerr.NewInternalError(rollbackErr)
		}
		return pgsqlc.CreateActorRow{}, domainErr
	}

	if _, err := store.GetMerchant(ctx, parsedMerchantID); err != nil {
		return rollbackOnErr(domainerr.MatchPostgresError(err))
	}

	var actorCount int
	countErr := tx.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM actors
        WHERE merchant_id = $1
    `, parsedMerchantID).Scan(&actorCount)
	if countErr != nil {
		return rollbackOnErr(domainerr.NewInternalError(countErr))
	}
	if actorCount > 0 {
		return rollbackOnErr(domainerr.NewDomainError(http.StatusConflict, domainerr.DuplicateError, "bootstrap actor already exists for merchant", fmt.Errorf("bootstrap actor already exists for merchant")))
	}

	for _, defaultRole := range []pgsqlc.RoleType{
		pgsqlc.RoleTypeMerchant,
		pgsqlc.RoleTypeEmployee,
		pgsqlc.RoleTypeCustomer,
	} {
		if _, roleErr := ensureMerchantRole(ctx, tx, store, parsedMerchantID, defaultRole); roleErr != nil {
			return rollbackOnErr(roleErr)
		}
	}

	created, createErr := store.CreateActor(ctx, pgsqlc.CreateActorParams{
		MerchantID:   parsedMerchantID,
		Email:        actor.Email,
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		IsActive:     true,
		LastLogin:    time.Now(),
	})
	if createErr != nil {
		return rollbackOnErr(domainerr.MatchPostgresError(createErr))
	}
	createdActor = created

	roleID, roleErr := ensureMerchantRole(ctx, tx, store, parsedMerchantID, roleType)
	if roleErr != nil {
		return rollbackOnErr(roleErr)
	}

	if _, assignErr := store.AssignActorRole(ctx, pgsqlc.AssignActorRoleParams{
		MerchantID: parsedMerchantID,
		ActorID:    createdActor.UID,
		RoleID:     roleID,
	}); assignErr != nil {
		return rollbackOnErr(domainerr.MatchPostgresError(assignErr))
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return pgsqlc.CreateActorRow{}, domainerr.NewInternalError(commitErr)
	}

	return createdActor, nil
}

func ensureMerchantRole(ctx context.Context, tx pgx.Tx, store *pgsqlc.Queries, merchantID uuid.UUID, roleType pgsqlc.RoleType) (uuid.UUID, *domainerr.DomainError) {
	var roleID uuid.UUID
	err := tx.QueryRow(ctx, `
        SELECT id
        FROM roles
        WHERE merchant_id = $1 AND role_type = $2
        LIMIT 1
    `, merchantID, roleType).Scan(&roleID)
	if err == nil {
		return roleID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, domainerr.NewInternalError(err)
	}

	role, createErr := store.CreateRole(ctx, pgsqlc.CreateRoleParams{
		MerchantID:  merchantID,
		RoleType:    roleType,
		Description: textValue(fmt.Sprintf("%s role", roleType)),
	})
	if createErr != nil {
		return uuid.Nil, domainerr.MatchPostgresError(createErr)
	}

	return role.ID, nil
}

func splitFullName(fullName string) (string, string) {
	nameParts := strings.Fields(strings.TrimSpace(fullName))
	firstName := ""
	lastName := ""
	if len(nameParts) > 0 {
		firstName = nameParts[0]
	}
	if len(nameParts) > 1 {
		lastName = strings.Join(nameParts[1:], " ")
	}
	return firstName, lastName
}
