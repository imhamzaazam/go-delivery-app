package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	actorstore2 "github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"
	"github.com/horiondreher/go-web-api-boilerplate/internal/auth/store"
	store2 "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
	"github.com/jackc/pgx/v5"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	merchant "github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) BootstrapActor(ctx context.Context, merchantID string, actor merchant.BootstrapActor) (actorstore2.CreateActorRow, *domainerr.DomainError) {
	if service.db == nil {
		return actorstore2.CreateActorRow{}, domainerr.NewInternalError(errors.New("merchant manager database is not configured"))
	}

	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return actorstore2.CreateActorRow{}, parseErr
	}

	roleType := store2.RoleType(strings.ToLower(strings.TrimSpace(actor.Role)))
	switch roleType {
	case store2.RoleTypeMerchant, store2.RoleTypeEmployee, store2.RoleTypeCustomer:
	default:
		return actorstore2.CreateActorRow{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid bootstrap role", fmt.Errorf("invalid bootstrap role"))
	}

	hashedPassword, hashErr := utils.HashPassword(actor.Password)
	if hashErr != nil {
		return actorstore2.CreateActorRow{}, hashErr
	}

	firstName, lastName := splitFullName(actor.FullName)
	var createdActor actorstore2.CreateActorRow

	tx, beginErr := service.db.BeginTx(ctx, pgx.TxOptions{})
	if beginErr != nil {
		return actorstore2.CreateActorRow{}, domainerr.NewInternalError(beginErr)
	}
	actorStore := actorstore2.New(tx)
	authStore := store.New(tx)
	merchantStore := store2.New(tx)

	rollbackOnErr := func(domainErr *domainerr.DomainError) (actorstore2.CreateActorRow, *domainerr.DomainError) {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return actorstore2.CreateActorRow{}, domainerr.NewInternalError(rollbackErr)
		}
		return actorstore2.CreateActorRow{}, domainErr
	}

	if _, err := merchantStore.GetMerchant(ctx, parsedMerchantID); err != nil {
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

	for _, defaultRole := range []store2.RoleType{
		store2.RoleTypeMerchant,
		store2.RoleTypeEmployee,
		store2.RoleTypeCustomer,
	} {
		if _, roleErr := ensureMerchantRole(ctx, tx, merchantStore, parsedMerchantID, defaultRole); roleErr != nil {
			return rollbackOnErr(roleErr)
		}
	}

	created, createErr := actorStore.CreateActor(ctx, actorstore2.CreateActorParams{
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

	roleID, roleErr := ensureMerchantRole(ctx, tx, merchantStore, parsedMerchantID, roleType)
	if roleErr != nil {
		return rollbackOnErr(roleErr)
	}

	if _, assignErr := authStore.AssignActorRole(ctx, store.AssignActorRoleParams{
		MerchantID: parsedMerchantID,
		ActorID:    createdActor.UID,
		RoleID:     roleID,
	}); assignErr != nil {
		return rollbackOnErr(domainerr.MatchPostgresError(assignErr))
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return actorstore2.CreateActorRow{}, domainerr.NewInternalError(commitErr)
	}

	return createdActor, nil
}

type roleCreator interface {
	CreateRole(ctx context.Context, arg store2.CreateRoleParams) (store2.Role, error)
}

func ensureMerchantRole(ctx context.Context, tx pgx.Tx, store roleCreator, merchantID uuid.UUID, roleType store2.RoleType) (uuid.UUID, *domainerr.DomainError) {
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

	role, createErr := store.CreateRole(ctx, store2.CreateRoleParams{
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
