package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	store2 "github.com/horiondreher/go-web-api-boilerplate/internal/auth/store"
	commercestore "github.com/horiondreher/go-web-api-boilerplate/internal/commerce/store"
	"github.com/jackc/pgx/v5"

	commerce "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	merchant "github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) CreateMerchantByAdmin(ctx context.Context, adminActorID uuid.UUID, newMerchant merchant.NewMerchant) (commerce.Merchant, *domainerr.DomainError) {
	ok, err := service.hasRole(ctx, adminActorID, commerce.RoleTypeAdmin, uuid.Nil)
	if err != nil {
		return commerce.Merchant{}, domainerr.NewInternalError(err)
	}
	if !ok {
		return commerce.Merchant{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "admin role required", fmt.Errorf("admin role required"))
	}

	var createdMerchant commerce.Merchant
	txErr := service.runInTx(ctx, func(_ pgx.Tx, store *commercestore.Postgres) *domainerr.DomainError {
		merchant, createErr := store.CreateMerchant(ctx, commerce.CreateMerchantParams{
			Name:          newMerchant.Name,
			Ntn:           newMerchant.Ntn,
			Address:       newMerchant.Address,
			Category:      commerce.MerchantCategory(strings.ToLower(newMerchant.Category)),
			ContactNumber: newMerchant.ContactNumber,
		})
		if createErr != nil {
			return domainerr.MatchPostgresError(createErr)
		}

		createdMerchant = merchant

		for _, roleType := range []commerce.RoleType{
			commerce.RoleTypeMerchant,
			commerce.RoleTypeEmployee,
			commerce.RoleTypeCustomer,
		} {
			if _, roleErr := store.CreateRole(ctx, commerce.CreateRoleParams{
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
		return commerce.Merchant{}, txErr
	}

	return createdMerchant, nil
}

func (service *Service) CreateEmployeeByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, fullName string, email string, password string) (commerce.CreateActorRow, *domainerr.DomainError) {
	allowed, allowErr := service.hasRole(ctx, merchantActorID, commerce.RoleTypeMerchant, merchantID)
	if allowErr != nil {
		return commerce.CreateActorRow{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return commerce.CreateActorRow{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "merchant role required", fmt.Errorf("merchant role required"))
	}

	hashedPassword, hashErr := utils.HashPassword(password)
	if hashErr != nil {
		return commerce.CreateActorRow{}, hashErr
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

	var createdActor commerce.CreateActorRow
	txErr := service.runInTx(ctx, func(tx pgx.Tx, store *commercestore.Postgres) *domainerr.DomainError {
		authStore := store2.New(tx)
		actor, createErr := store.CreateActor(ctx, commerce.CreateActorParams{
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
		`, merchantID, commerce.RoleTypeEmployee).Scan(&employeeRoleID)
		if roleErr != nil {
			return domainerr.NewInternalError(roleErr)
		}

		if _, assignErr := authStore.AssignActorRole(ctx, store2.AssignActorRoleParams{
			MerchantID: merchantID,
			ActorID:    createdActor.UID,
			RoleID:     employeeRoleID,
		}); assignErr != nil {
			return domainerr.MatchPostgresError(assignErr)
		}

		return nil
	})
	if txErr != nil {
		return commerce.CreateActorRow{}, txErr
	}

	return createdActor, nil
}
