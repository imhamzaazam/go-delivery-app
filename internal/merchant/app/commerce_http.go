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
	"github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
	"github.com/jackc/pgx/v5"

	actor "github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	commerce "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	merchant "github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) CreateActorByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, newActor actor.NewActor, role string) (actor.CreateActorRow, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return actor.CreateActorRow{}, parseErr
	}

	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return actor.CreateActorRow{}, accessErr
	}

	roleType := store.RoleType(strings.ToLower(strings.TrimSpace(role)))
	switch roleType {
	case store.RoleTypeMerchant, store.RoleTypeEmployee, store.RoleTypeCustomer:
	default:
		return actor.CreateActorRow{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid actor role", fmt.Errorf("invalid actor role"))
	}

	hashedPassword, hashErr := utils.HashPasswordOrNoop(newActor.Password)
	if hashErr != nil {
		return actor.CreateActorRow{}, hashErr
	}

	firstName, lastName := splitFullName(newActor.FullName)
	var createdActor actor.CreateActorRow

	txErr := service.runInTx(ctx, func(tx pgx.Tx, store *commercestore.Postgres) *domainerr.DomainError {
		authQueries := store2.New(tx)
		merchantQueries := store.New(tx)

		created, createErr := store.CreateActor(ctx, commerce.CreateActorParams{
			MerchantID:   parsedMerchantID,
			Email:        newActor.Email,
			PasswordHash: hashedPassword,
			FirstName:    firstName,
			LastName:     lastName,
			IsActive:     true,
			LastLogin:    time.Now(),
		})
		if createErr != nil {
			return domainerr.MatchPostgresError(createErr)
		}
		createdActor = created

		roleID, roleErr := ensureMerchantRole(ctx, tx, merchantQueries, parsedMerchantID, roleType)
		if roleErr != nil {
			return roleErr
		}

		if _, assignErr := authQueries.AssignActorRole(ctx, store2.AssignActorRoleParams{
			MerchantID: parsedMerchantID,
			ActorID:    createdActor.UID,
			RoleID:     roleID,
		}); assignErr != nil {
			return domainerr.MatchPostgresError(assignErr)
		}

		return nil
	})
	if txErr != nil {
		return actor.CreateActorRow{}, txErr
	}

	return createdActor, nil
}

func (service *Service) CreateBranchByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, name string, address string, contactNumber string, city string, openingTime string, closingTime string) (merchant.Branch, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return merchant.Branch{}, parseErr
	}

	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return merchant.Branch{}, accessErr
	}

	openingMinutes, openingErr := utils.ParseBranchClock(openingTime)
	if openingErr != nil {
		return merchant.Branch{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid opening time", openingErr)
	}
	closingMinutes, closingErr := utils.ParseBranchClock(closingTime)
	if closingErr != nil {
		return merchant.Branch{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid closing time", closingErr)
	}

	branch, err := service.store.CreateBranch(ctx, commerce.CreateBranchParams{
		MerchantID:         parsedMerchantID,
		Name:               name,
		Address:            address,
		ContactNumber:      textValue(contactNumber),
		City:               commerce.CityType(city),
		OpeningTimeMinutes: openingMinutes,
		ClosingTimeMinutes: closingMinutes,
	})
	if err != nil {
		return merchant.Branch{}, domainerr.MatchPostgresError(err)
	}

	return branch, nil
}

func (service *Service) GetBranchAvailability(ctx context.Context, merchantID string, branchID string) (merchant.BranchAvailability, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := utils.ParseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return merchant.BranchAvailability{}, merchantErr
	}
	parsedBranchID, branchErr := utils.ParseUUID(branchID, "branch id")
	if branchErr != nil {
		return merchant.BranchAvailability{}, branchErr
	}

	branch, err := service.store.GetBranch(ctx, commerce.GetBranchParams{
		MerchantID: parsedMerchantID,
		ID:         parsedBranchID,
	})
	if err != nil {
		return merchant.BranchAvailability{}, domainerr.MatchPostgresError(err)
	}

	now := time.Now().In(utils.BranchLocation())
	return merchant.BranchAvailability{
		Branch:       branch,
		IsOpen:       utils.IsBranchOpenAt(branch.OpeningTimeMinutes, branch.ClosingTimeMinutes, now),
		OpeningTime:  utils.FormatBranchClock(branch.OpeningTimeMinutes),
		ClosingTime:  utils.FormatBranchClock(branch.ClosingTimeMinutes),
		CurrentTime:  now.Format("15:04"),
		TimezoneName: utils.RestaurantTimezone,
	}, nil
}

func (service *Service) CreateMerchantDiscountByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, discountType string, value float64, description string, productID string, categoryID string) (merchant.MerchantDiscount, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return merchant.MerchantDiscount{}, parseErr
	}

	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return merchant.MerchantDiscount{}, accessErr
	}

	parsedDiscountType := store.DiscountType(strings.ToLower(strings.TrimSpace(discountType)))
	switch parsedDiscountType {
	case store.DiscountTypeFlat, store.DiscountTypePercentage:
	default:
		return merchant.MerchantDiscount{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid discount type", fmt.Errorf("invalid discount type"))
	}

	parsedProductID := uuid.Nil
	if strings.TrimSpace(productID) != "" {
		value, productErr := utils.ParseUUID(productID, "product id")
		if productErr != nil {
			return merchant.MerchantDiscount{}, productErr
		}
		if _, err := service.store.GetProduct(ctx, commerce.GetProductParams{MerchantID: parsedMerchantID, ID: value}); err != nil {
			return merchant.MerchantDiscount{}, domainerr.MatchPostgresError(err)
		}
		parsedProductID = value
	}

	parsedCategoryID := uuid.Nil
	if strings.TrimSpace(categoryID) != "" {
		value, categoryErr := utils.ParseUUID(categoryID, "category id")
		if categoryErr != nil {
			return merchant.MerchantDiscount{}, categoryErr
		}
		if _, err := service.store.GetProductCategory(ctx, commerce.GetProductCategoryParams{MerchantID: parsedMerchantID, ID: value}); err != nil {
			return merchant.MerchantDiscount{}, domainerr.MatchPostgresError(err)
		}
		parsedCategoryID = value
	}

	createdDiscount, err := service.store.CreateMerchantDiscount(ctx, commerce.CreateMerchantDiscountParams{
		MerchantID:  parsedMerchantID,
		ProductID:   parsedProductID,
		CategoryID:  parsedCategoryID,
		Type:        parsedDiscountType,
		Value:       utils.NumericFromFloat(value),
		Description: textValue(description),
	})
	if err != nil {
		return merchant.MerchantDiscount{}, domainerr.MatchPostgresError(err)
	}

	return merchantDiscountFromCreateRow(createdDiscount), nil
}
