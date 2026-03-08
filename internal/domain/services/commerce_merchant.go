package services

import (
	"context"
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

func (service *CommerceManager) CreateMerchantByAdmin(ctx context.Context, adminActorID uuid.UUID, newMerchant ports.NewMerchant) (pgsqlc.Merchant, *domainerr.DomainError) {
	ok, err := service.hasRole(ctx, adminActorID, pgsqlc.RoleTypeAdmin, uuid.Nil)
	if err != nil {
		return pgsqlc.Merchant{}, domainerr.NewInternalError(err)
	}
	if !ok {
		return pgsqlc.Merchant{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "admin role required", fmt.Errorf("admin role required"))
	}

	var createdMerchant pgsqlc.Merchant
	txErr := service.runInTx(ctx, func(_ pgx.Tx, store *pgsqlc.Queries) *domainerr.DomainError {
		merchant, createErr := store.CreateMerchant(ctx, pgsqlc.CreateMerchantParams{
			Name:          newMerchant.Name,
			Ntn:           newMerchant.Ntn,
			Address:       newMerchant.Address,
			Category:      pgsqlc.MerchantCategory(strings.ToLower(newMerchant.Category)),
			ContactNumber: newMerchant.ContactNumber,
		})
		if createErr != nil {
			return domainerr.MatchPostgresError(createErr)
		}

		createdMerchant = merchant

		for _, roleType := range []pgsqlc.RoleType{
			pgsqlc.RoleTypeMerchant,
			pgsqlc.RoleTypeEmployee,
			pgsqlc.RoleTypeCustomer,
		} {
			if _, roleErr := store.CreateRole(ctx, pgsqlc.CreateRoleParams{
				MerchantID:  createdMerchant.ID,
				RoleType:    roleType,
				Description: textValue(fmt.Sprintf("%s role", roleType)),
			}); roleErr != nil {
				return domainerr.MatchPostgresError(roleErr)
			}
		}

		return nil
	})
	if txErr != nil {
		return pgsqlc.Merchant{}, txErr
	}

	return createdMerchant, nil
}

func (service *CommerceManager) CreateEmployeeByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, fullName string, email string, password string) (pgsqlc.CreateActorRow, *domainerr.DomainError) {
	allowed, allowErr := service.hasRole(ctx, merchantActorID, pgsqlc.RoleTypeMerchant, merchantID)
	if allowErr != nil {
		return pgsqlc.CreateActorRow{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return pgsqlc.CreateActorRow{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "merchant role required", fmt.Errorf("merchant role required"))
	}

	hashedPassword, hashErr := utils.HashPassword(password)
	if hashErr != nil {
		return pgsqlc.CreateActorRow{}, hashErr
	}

	nameParts := strings.Fields(strings.TrimSpace(fullName))
	firstName := ""
	lastName := ""
	if len(nameParts) > 0 {
		firstName = nameParts[0]
	}
	if len(nameParts) > 1 {
		lastName = strings.Join(nameParts[1:], " ")
	}

	var createdActor pgsqlc.CreateActorRow
	txErr := service.runInTx(ctx, func(tx pgx.Tx, store *pgsqlc.Queries) *domainerr.DomainError {
		actor, createErr := store.CreateActor(ctx, pgsqlc.CreateActorParams{
			MerchantID:   merchantID,
			Email:        email,
			PasswordHash: hashedPassword,
			FirstName:    firstName,
			LastName:     lastName,
			IsActive:     true,
			LastLogin:    time.Now(),
		})
		if createErr != nil {
			return domainerr.MatchPostgresError(createErr)
		}
		createdActor = actor

		var employeeRoleID uuid.UUID
		roleErr := tx.QueryRow(ctx, `
        SELECT id
        FROM roles
        WHERE merchant_id = $1 AND role_type = $2
        LIMIT 1
    `, merchantID, pgsqlc.RoleTypeEmployee).Scan(&employeeRoleID)
		if roleErr != nil {
			return domainerr.NewInternalError(roleErr)
		}

		if _, assignErr := store.AssignActorRole(ctx, pgsqlc.AssignActorRoleParams{
			MerchantID: merchantID,
			ActorID:    createdActor.UID,
			RoleID:     employeeRoleID,
		}); assignErr != nil {
			return domainerr.MatchPostgresError(assignErr)
		}

		return nil
	})
	if txErr != nil {
		return pgsqlc.CreateActorRow{}, txErr
	}

	return createdActor, nil
}
