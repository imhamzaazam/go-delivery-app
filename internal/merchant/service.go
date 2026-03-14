package merchant

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	actor "github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	actorpostgres "github.com/horiondreher/go-web-api-boilerplate/internal/actor/store"
	actorstore "github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"
	authstore "github.com/horiondreher/go-web-api-boilerplate/internal/auth/store"
	catalog "github.com/horiondreher/go-web-api-boilerplate/internal/catalog"
	catalogstore "github.com/horiondreher/go-web-api-boilerplate/internal/catalog/store"
	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	pkgdb "github.com/horiondreher/go-web-api-boilerplate/pkg/db"
)

type MerchantService struct {
	db           *pkgdb.DB
	store        *merchantstore.Postgres
	actorStore   *actorpostgres.Postgres
	catalogStore *catalogstore.Postgres
}

func NewService(db *pkgdb.DB, store *merchantstore.Postgres) *MerchantService {
	return &MerchantService{
		db:           db,
		store:        store,
		actorStore:   actorpostgres.New(db.Pool),
		catalogStore: catalogstore.New(db.Pool),
	}
}

type merchantQueryStore interface {
	CreateMerchant(ctx context.Context, arg merchantstore.CreateMerchantParams) (Merchant, error)
	UpdateMerchant(ctx context.Context, arg merchantstore.UpdateMerchantParams) (Merchant, error)
	GetMerchant(ctx context.Context, id uuid.UUID) (Merchant, error)
	ListMerchants(ctx context.Context) ([]Merchant, error)
	ListBranchesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]merchantstore.ListBranchesByMerchantRow, error)
	ListDiscountsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]merchantstore.ListDiscountsByMerchantRow, error)
	ListRolesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Role, error)
}

type MerchantManager struct {
	db    *pkgdb.DB
	store merchantQueryStore
}

func NewMerchantManager(db *pkgdb.DB, store merchantQueryStore) *MerchantManager {
	return &MerchantManager{
		db:    db,
		store: store,
	}
}

func (service *MerchantService) runInTx(ctx context.Context, fn func(tx pgx.Tx) *domainerr.DomainError) *domainerr.DomainError {
	tx, err := service.db.Pool.Begin(ctx)
	if err != nil {
		return domainerr.NewInternalError(err)
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return domainerr.NewInternalError(err)
	}
	return nil
}

func (service *MerchantService) CreateMerchant(ctx context.Context, newMerchant NewMerchant) (Merchant, *domainerr.DomainError) {
	category := merchantstore.MerchantCategory(strings.ToLower(newMerchant.Category))

	createdMerchant, err := service.store.CreateMerchant(ctx, merchantstore.CreateMerchantParams{
		Name:          newMerchant.Name,
		Ntn:           newMerchant.Ntn,
		Address:       newMerchant.Address,
		Category:      category,
		ContactNumber: newMerchant.ContactNumber,
	})
	if err != nil {
		return Merchant{}, domainerr.MatchPostgresError(err)
	}

	return mapMerchant(createdMerchant), nil
}

func (service *MerchantService) UpdateMerchant(ctx context.Context, merchantID string, newMerchant NewMerchant) (Merchant, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return Merchant{}, parseErr
	}

	updatedMerchant, err := service.store.UpdateMerchant(ctx, merchantstore.UpdateMerchantParams{
		ID:            parsedMerchantID,
		Name:          newMerchant.Name,
		Ntn:           newMerchant.Ntn,
		Address:       newMerchant.Address,
		Category:      merchantstore.MerchantCategory(strings.ToLower(newMerchant.Category)),
		ContactNumber: newMerchant.ContactNumber,
	})
	if err != nil {
		return Merchant{}, domainerr.MatchPostgresError(err)
	}

	return mapMerchant(updatedMerchant), nil
}

func (service *MerchantService) GetMerchant(ctx context.Context, merchantID string) (Merchant, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return Merchant{}, parseErr
	}

	merchantRow, err := service.store.GetMerchant(ctx, parsedMerchantID)
	if err != nil {
		return Merchant{}, domainerr.MatchPostgresError(err)
	}

	return mapMerchant(merchantRow), nil
}

func (service *MerchantService) ListMerchants(ctx context.Context) ([]Merchant, *domainerr.DomainError) {
	merchants, err := service.store.ListMerchants(ctx)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	result := make([]Merchant, 0, len(merchants))
	for _, m := range merchants {
		result = append(result, mapMerchant(m))
	}

	return result, nil
}

func (service *MerchantService) ListBranchesByMerchant(ctx context.Context, merchantID string) ([]Branch, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	branches, err := service.store.ListBranchesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	result := make([]Branch, 0, len(branches))
	for _, branch := range branches {
		result = append(result, mapBranch(merchantstore.Branch{ID: branch.ID, MerchantID: branch.MerchantID, Name: branch.Name, Address: branch.Address, ContactNumber: branch.ContactNumber, City: branch.City, OpeningTimeMinutes: branch.OpeningTimeMinutes, ClosingTimeMinutes: branch.ClosingTimeMinutes, CreatedAt: branch.CreatedAt, UpdatedAt: branch.UpdatedAt}))
	}

	return result, nil
}

func (service *MerchantService) ListDiscountsByMerchant(ctx context.Context, merchantID string) ([]MerchantDiscount, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	discounts, err := service.store.ListDiscountsByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	result := make([]MerchantDiscount, 0, len(discounts))
	for _, discount := range discounts {
		result = append(result, mapMerchantDiscount(merchantstore.MerchantDiscount{ID: discount.ID, MerchantID: discount.MerchantID, Type: discount.Type, Value: discount.Value, Description: discount.Description, ValidFrom: discount.ValidFrom, ValidTo: discount.ValidTo, CreatedAt: discount.CreatedAt, ProductID: discount.ProductID, CategoryID: discount.CategoryID}))
	}

	return result, nil
}

func (service *MerchantService) ListRolesByMerchant(ctx context.Context, merchantID string) ([]Role, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return nil, parseErr
	}

	roles, err := service.store.ListRolesByMerchant(ctx, parsedMerchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	result := make([]Role, 0, len(roles))
	for _, r := range roles {
		result = append(result, mapRole(r))
	}

	return result, nil
}

func (service *MerchantService) resolveViewerActor(ctx context.Context, merchantID uuid.UUID, email string) (actorstore.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
	viewer, err := service.actorStore.GetActorProfileByMerchantAndEmail(ctx, actorstore.GetActorProfileByMerchantAndEmailParams{
		MerchantID: merchantID,
		Email:      email,
	})
	if err != nil {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, domainerr.MatchPostgresError(err)
	}

	return viewer, nil
}

func (service *MerchantService) requireMerchantViewAccess(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, targetMerchantID uuid.UUID) (actorstore.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
	viewer, viewerErr := service.resolveViewerActor(ctx, viewerMerchantID, viewerEmail)
	if viewerErr != nil {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, viewerErr
	}

	canView, allowErr := service.canViewMerchant(ctx, viewer.UID, targetMerchantID)
	if allowErr != nil {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewInternalError(allowErr)
	}
	if !canView {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "not allowed to view merchant", fmt.Errorf("not allowed to view merchant"))
	}

	return viewer, nil
}

func (service *MerchantService) requireMerchantManageAccess(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, targetMerchantID uuid.UUID) (actorstore.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
	viewer, viewerErr := service.resolveViewerActor(ctx, viewerMerchantID, viewerEmail)
	if viewerErr != nil {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, viewerErr
	}

	isAdmin, adminErr := service.hasRole(ctx, viewer.UID, merchantstore.RoleTypeAdmin, uuid.Nil)
	if adminErr != nil {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewInternalError(adminErr)
	}
	if isAdmin {
		return viewer, nil
	}

	isMerchant, merchantErr := service.hasRole(ctx, viewer.UID, merchantstore.RoleTypeMerchant, targetMerchantID)
	if merchantErr != nil {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewInternalError(merchantErr)
	}
	if !isMerchant {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "merchant role required", fmt.Errorf("merchant role required"))
	}

	return viewer, nil
}

func (service *MerchantService) canViewMerchant(ctx context.Context, actorID uuid.UUID, merchantID uuid.UUID) (bool, error) {
	isAdmin, adminErr := service.hasRole(ctx, actorID, merchantstore.RoleTypeAdmin, uuid.Nil)
	if adminErr != nil {
		return false, adminErr
	}
	if isAdmin {
		return true, nil
	}

	isMerchant, merchantErr := service.hasRole(ctx, actorID, merchantstore.RoleTypeMerchant, merchantID)
	if merchantErr != nil {
		return false, merchantErr
	}

	return isMerchant, nil
}

func (service *MerchantService) CreateMerchantByAdmin(ctx context.Context, adminActorID uuid.UUID, newMerchant NewMerchant) (Merchant, *domainerr.DomainError) {
	ok, err := service.hasRole(ctx, adminActorID, merchantstore.RoleTypeAdmin, uuid.Nil)
	if err != nil {
		return Merchant{}, domainerr.NewInternalError(err)
	}
	if !ok {
		return Merchant{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "admin role required", fmt.Errorf("admin role required"))
	}

	var createdMerchant Merchant
	txErr := service.runInTx(ctx, func(tx pgx.Tx) *domainerr.DomainError {
		merchantQueries := merchantstore.New(tx)
		created, createErr := merchantQueries.CreateMerchant(ctx, merchantstore.CreateMerchantParams{
			Name:          newMerchant.Name,
			Ntn:           newMerchant.Ntn,
			Address:       newMerchant.Address,
			Category:      merchantstore.MerchantCategory(strings.ToLower(newMerchant.Category)),
			ContactNumber: newMerchant.ContactNumber,
		})
		if createErr != nil {
			return domainerr.MatchPostgresError(createErr)
		}

		createdMerchant = mapMerchant(created)

		for _, roleType := range []merchantstore.RoleType{
			merchantstore.RoleTypeMerchant,
			merchantstore.RoleTypeEmployee,
			merchantstore.RoleTypeCustomer,
		} {
			if _, roleErr := merchantQueries.CreateRole(ctx, merchantstore.CreateRoleParams{
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
		return Merchant{}, txErr
	}

	return createdMerchant, nil
}

func (service *MerchantService) CreateEmployeeByMerchant(ctx context.Context, merchantActorID uuid.UUID, merchantID uuid.UUID, fullName string, email string, password string) (actor.CreateActorRow, *domainerr.DomainError) {
	allowed, allowErr := service.hasRole(ctx, merchantActorID, merchantstore.RoleTypeMerchant, merchantID)
	if allowErr != nil {
		return actor.CreateActorRow{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return actor.CreateActorRow{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "merchant role required", fmt.Errorf("merchant role required"))
	}

	hashedPassword, hashErr := utils.HashPassword(password)
	if hashErr != nil {
		return actor.CreateActorRow{}, hashErr
	}

	firstName, lastName := splitFullName(fullName)
	var createdActor actor.CreateActorRow
	txErr := service.runInTx(ctx, func(tx pgx.Tx) *domainerr.DomainError {
		actorQueries := actorstore.New(tx)
		authQueries := authstore.New(tx)
		created, createErr := actorQueries.CreateActor(ctx, actorstore.CreateActorParams{
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
		createdActor = created

		var employeeRoleID uuid.UUID
		roleErr := tx.QueryRow(ctx, `
        SELECT id
        FROM roles
        WHERE merchant_id = $1 AND role_type = $2
        LIMIT 1
		`, merchantID, merchantstore.RoleTypeEmployee).Scan(&employeeRoleID)
		if roleErr != nil {
			return domainerr.NewInternalError(roleErr)
		}

		if _, assignErr := authQueries.AssignActorRole(ctx, authstore.AssignActorRoleParams{
			MerchantID: merchantID,
			ActorID:    createdActor.UID,
			RoleID:     employeeRoleID,
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

func (service *MerchantService) CreateActorByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, newActor actor.NewActor, role string) (actor.CreateActorRow, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return actor.CreateActorRow{}, parseErr
	}

	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return actor.CreateActorRow{}, accessErr
	}

	roleType := merchantstore.RoleType(strings.ToLower(strings.TrimSpace(role)))
	switch roleType {
	case merchantstore.RoleTypeMerchant, merchantstore.RoleTypeEmployee, merchantstore.RoleTypeCustomer:
	default:
		return actor.CreateActorRow{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid actor role", fmt.Errorf("invalid actor role"))
	}

	hashedPassword, hashErr := utils.HashPasswordOrNoop(newActor.Password)
	if hashErr != nil {
		return actor.CreateActorRow{}, hashErr
	}

	firstName, lastName := splitFullName(newActor.FullName)
	var createdActor actor.CreateActorRow
	txErr := service.runInTx(ctx, func(tx pgx.Tx) *domainerr.DomainError {
		actorQueries := actorstore.New(tx)
		authQueries := authstore.New(tx)
		merchantQueries := merchantstore.New(tx)

		created, createErr := actorQueries.CreateActor(ctx, actorstore.CreateActorParams{
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

		if _, assignErr := authQueries.AssignActorRole(ctx, authstore.AssignActorRoleParams{
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

func (service *MerchantService) CreateBranchByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, name string, address string, contactNumber string, city string, openingTime string, closingTime string) (Branch, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return Branch{}, parseErr
	}

	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return Branch{}, accessErr
	}

	openingMinutes, openingErr := utils.ParseBranchClock(openingTime)
	if openingErr != nil {
		return Branch{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid opening time", openingErr)
	}
	closingMinutes, closingErr := utils.ParseBranchClock(closingTime)
	if closingErr != nil {
		return Branch{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid closing time", closingErr)
	}

	branch, err := service.store.CreateBranch(ctx, merchantstore.CreateBranchParams{
		MerchantID:         parsedMerchantID,
		Name:               name,
		Address:            address,
		ContactNumber:      textValue(contactNumber),
		City:               merchantstore.CityType(city),
		OpeningTimeMinutes: openingMinutes,
		ClosingTimeMinutes: closingMinutes,
	})
	if err != nil {
		return Branch{}, domainerr.MatchPostgresError(err)
	}

	return mapBranch(branch), nil
}

func (service *MerchantService) GetBranchAvailability(ctx context.Context, merchantID string, branchID string) (BranchAvailability, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := utils.ParseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return BranchAvailability{}, merchantErr
	}
	parsedBranchID, branchErr := utils.ParseUUID(branchID, "branch id")
	if branchErr != nil {
		return BranchAvailability{}, branchErr
	}

	branch, err := service.store.GetBranch(ctx, merchantstore.GetBranchParams{
		MerchantID: parsedMerchantID,
		ID:         parsedBranchID,
	})
	if err != nil {
		return BranchAvailability{}, domainerr.MatchPostgresError(err)
	}

	now := time.Now().In(utils.BranchLocation())
	return BranchAvailability{
		Branch:       mapBranch(branch),
		IsOpen:       utils.IsBranchOpenAt(branch.OpeningTimeMinutes, branch.ClosingTimeMinutes, now),
		OpeningTime:  utils.FormatBranchClock(branch.OpeningTimeMinutes),
		ClosingTime:  utils.FormatBranchClock(branch.ClosingTimeMinutes),
		CurrentTime:  now.Format("15:04"),
		TimezoneName: utils.RestaurantTimezone,
	}, nil
}

func (service *MerchantService) CreateMerchantDiscountByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, discountType string, value float64, description string, productID string, categoryID string) (MerchantDiscount, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return MerchantDiscount{}, parseErr
	}

	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return MerchantDiscount{}, accessErr
	}

	parsedDiscountType := merchantstore.DiscountType(strings.ToLower(strings.TrimSpace(discountType)))
	switch parsedDiscountType {
	case merchantstore.DiscountTypeFlat, merchantstore.DiscountTypePercentage:
	default:
		return MerchantDiscount{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid discount type", fmt.Errorf("invalid discount type"))
	}

	parsedProductID := uuid.Nil
	if strings.TrimSpace(productID) != "" {
		value, productErr := utils.ParseUUID(productID, "product id")
		if productErr != nil {
			return MerchantDiscount{}, productErr
		}
		if _, err := service.catalogStore.GetProduct(ctx, catalog.GetProductParams{MerchantID: parsedMerchantID, ID: value}); err != nil {
			return MerchantDiscount{}, domainerr.MatchPostgresError(err)
		}
		parsedProductID = value
	}

	parsedCategoryID := uuid.Nil
	if strings.TrimSpace(categoryID) != "" {
		value, categoryErr := utils.ParseUUID(categoryID, "category id")
		if categoryErr != nil {
			return MerchantDiscount{}, categoryErr
		}
		if _, err := service.catalogStore.GetProductCategory(ctx, catalog.GetProductCategoryParams{MerchantID: parsedMerchantID, ID: value}); err != nil {
			return MerchantDiscount{}, domainerr.MatchPostgresError(err)
		}
		parsedCategoryID = value
	}

	createdDiscount, err := service.store.CreateMerchantDiscount(ctx, merchantstore.CreateMerchantDiscountParams{
		MerchantID:  parsedMerchantID,
		ProductID:   parsedProductID,
		CategoryID:  parsedCategoryID,
		Type:        parsedDiscountType,
		Value:       utils.NumericFromFloat(value),
		Description: textValue(description),
	})
	if err != nil {
		return MerchantDiscount{}, domainerr.MatchPostgresError(err)
	}

	return mapMerchantDiscount(merchantstore.MerchantDiscount{
		ID:          createdDiscount.ID,
		MerchantID:  createdDiscount.MerchantID,
		Type:        createdDiscount.Type,
		Value:       createdDiscount.Value,
		Description: createdDiscount.Description,
		ValidFrom:   createdDiscount.ValidFrom,
		ValidTo:     createdDiscount.ValidTo,
		CreatedAt:   createdDiscount.CreatedAt,
		ProductID:   createdDiscount.ProductID,
		CategoryID:  createdDiscount.CategoryID,
	}), nil
}

func (service *MerchantService) BootstrapActor(ctx context.Context, merchantID string, actor BootstrapActor) (actorstore.CreateActorRow, *domainerr.DomainError) {
	if service.db == nil {
		return actorstore.CreateActorRow{}, domainerr.NewInternalError(errors.New("merchant manager database is not configured"))
	}

	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return actorstore.CreateActorRow{}, parseErr
	}

	roleType := merchantstore.RoleType(strings.ToLower(strings.TrimSpace(actor.Role)))
	switch roleType {
	case merchantstore.RoleTypeMerchant, merchantstore.RoleTypeEmployee, merchantstore.RoleTypeCustomer:
	default:
		return actorstore.CreateActorRow{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid bootstrap role", fmt.Errorf("invalid bootstrap role"))
	}

	hashedPassword, hashErr := utils.HashPassword(actor.Password)
	if hashErr != nil {
		return actorstore.CreateActorRow{}, hashErr
	}

	firstName, lastName := splitFullName(actor.FullName)
	var createdActor actorstore.CreateActorRow

	tx, beginErr := service.db.BeginTx(ctx, pgx.TxOptions{})
	if beginErr != nil {
		return actorstore.CreateActorRow{}, domainerr.NewInternalError(beginErr)
	}
	actorQueries := actorstore.New(tx)
	authQueries := authstore.New(tx)
	merchantQueries := merchantstore.New(tx)

	rollbackOnErr := func(domainErr *domainerr.DomainError) (actorstore.CreateActorRow, *domainerr.DomainError) {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return actorstore.CreateActorRow{}, domainerr.NewInternalError(rollbackErr)
		}
		return actorstore.CreateActorRow{}, domainErr
	}

	if _, err := merchantQueries.GetMerchant(ctx, parsedMerchantID); err != nil {
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

	for _, defaultRole := range []merchantstore.RoleType{
		merchantstore.RoleTypeMerchant,
		merchantstore.RoleTypeEmployee,
		merchantstore.RoleTypeCustomer,
	} {
		if _, roleErr := ensureMerchantRole(ctx, tx, merchantQueries, parsedMerchantID, defaultRole); roleErr != nil {
			return rollbackOnErr(roleErr)
		}
	}

	created, createErr := actorQueries.CreateActor(ctx, actorstore.CreateActorParams{
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

	roleID, roleErr := ensureMerchantRole(ctx, tx, merchantQueries, parsedMerchantID, roleType)
	if roleErr != nil {
		return rollbackOnErr(roleErr)
	}

	if _, assignErr := authQueries.AssignActorRole(ctx, authstore.AssignActorRoleParams{
		MerchantID: parsedMerchantID,
		ActorID:    createdActor.UID,
		RoleID:     roleID,
	}); assignErr != nil {
		return rollbackOnErr(domainerr.MatchPostgresError(assignErr))
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return actorstore.CreateActorRow{}, domainerr.NewInternalError(commitErr)
	}

	return createdActor, nil
}

type roleCreator interface {
	CreateRole(ctx context.Context, arg merchantstore.CreateRoleParams) (merchantstore.Role, error)
}

func ensureMerchantRole(ctx context.Context, tx pgx.Tx, store roleCreator, merchantID uuid.UUID, roleType merchantstore.RoleType) (uuid.UUID, *domainerr.DomainError) {
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

	role, createErr := store.CreateRole(ctx, merchantstore.CreateRoleParams{
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

func (service *MerchantService) hasRole(ctx context.Context, actorID uuid.UUID, roleType merchantstore.RoleType, merchantID uuid.UUID) (bool, error) {
	var one int
	var err error
	if merchantID == uuid.Nil {
		err = service.db.QueryRow(ctx, `
            SELECT 1
            FROM actors a
            JOIN actor_roles ar ON ar.actor_id = a.id AND ar.merchant_id = a.merchant_id
            JOIN roles r ON r.id = ar.role_id AND r.merchant_id = a.merchant_id
            WHERE a.id = $1 AND r.role_type = $2
            LIMIT 1
        `, actorID, roleType).Scan(&one)
	} else {
		err = service.db.QueryRow(ctx, `
            SELECT 1
            FROM actors a
            JOIN actor_roles ar ON ar.actor_id = a.id AND ar.merchant_id = a.merchant_id
            JOIN roles r ON r.id = ar.role_id AND r.merchant_id = a.merchant_id
            WHERE a.id = $1 AND a.merchant_id = $2 AND r.role_type = $3
            LIMIT 1
        `, actorID, merchantID, roleType).Scan(&one)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (service *MerchantService) getRoleIDByType(ctx context.Context, merchantID uuid.UUID, roleType merchantstore.RoleType) (uuid.UUID, error) {
	var roleID uuid.UUID
	err := service.db.QueryRow(ctx, `
        SELECT id
        FROM roles
        WHERE merchant_id = $1 AND role_type = $2
        LIMIT 1
    `, merchantID, roleType).Scan(&roleID)
	if err != nil {
		return uuid.Nil, err
	}

	return roleID, nil
}

func textValue(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{String: value, Valid: true}
}

func isDiscountActive(discount merchantstore.MerchantDiscount, now time.Time) bool {
	if !discount.ValidFrom.IsZero() && now.Before(discount.ValidFrom) {
		return false
	}
	if !discount.ValidTo.IsZero() && now.After(discount.ValidTo) {
		return false
	}

	return true
}

const (
	productDiscountPriority = iota + 1
	categoryDiscountPriority
	merchantDiscountPriority
	noDiscountPriority
)

func discountPriority(discount merchantstore.MerchantDiscount, productID uuid.UUID, categoryID uuid.UUID) (int, bool) {
	switch {
	case discount.ProductID != uuid.Nil:
		return productDiscountPriority, discount.ProductID == productID
	case discount.CategoryID != uuid.Nil:
		return categoryDiscountPriority, discount.CategoryID == categoryID
	default:
		return merchantDiscountPriority, true
	}
}

func (service *MerchantService) resolveBestDiscountForProduct(ctx context.Context, merchantID uuid.UUID, productID uuid.UUID, categoryID uuid.UUID, now time.Time) (*merchantstore.MerchantDiscount, *domainerr.DomainError) {
	discounts, err := service.store.ListDiscountsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	var best *merchantstore.MerchantDiscount
	bestPriority := noDiscountPriority
	for _, row := range discounts {
		discount := merchantstore.MerchantDiscount{
			ID:          row.ID,
			MerchantID:  row.MerchantID,
			Type:        row.Type,
			Value:       row.Value,
			Description: row.Description,
			ValidFrom:   row.ValidFrom,
			ValidTo:     row.ValidTo,
			CreatedAt:   row.CreatedAt,
			ProductID:   row.ProductID,
			CategoryID:  row.CategoryID,
		}
		if !isDiscountActive(discount, now) {
			continue
		}

		priority, matches := discountPriority(discount, productID, categoryID)
		if !matches {
			continue
		}

		if best == nil || priority < bestPriority || (priority == bestPriority && discount.CreatedAt.After(best.CreatedAt)) {
			selected := discount
			best = &selected
			bestPriority = priority
		}
	}

	return best, nil
}

func mapMerchant(m merchantstore.Merchant) Merchant {
	var logo *string
	if m.Logo.Valid {
		logo = &m.Logo.String
	}
	return Merchant{
		ID:            m.ID,
		Name:          m.Name,
		Ntn:           m.Ntn,
		Address:       m.Address,
		Logo:          logo,
		Category:      MerchantCategory(m.Category),
		ContactNumber: m.ContactNumber,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func mapBranch(b merchantstore.Branch) Branch {
	var contact *string
	if b.ContactNumber.Valid {
		contact = &b.ContactNumber.String
	}
	return Branch{
		ID:                 b.ID,
		MerchantID:         b.MerchantID,
		Name:               b.Name,
		Address:            b.Address,
		ContactNumber:      contact,
		City:               CityType(b.City),
		OpeningTimeMinutes: b.OpeningTimeMinutes,
		ClosingTimeMinutes: b.ClosingTimeMinutes,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}
}

func mapMerchantDiscount(d merchantstore.MerchantDiscount) MerchantDiscount {
	var desc *string
	if d.Description.Valid {
		desc = &d.Description.String
	}
	val, _ := d.Value.Float64Value()
	return MerchantDiscount{
		ID:          d.ID,
		MerchantID:  d.MerchantID,
		Type:        DiscountType(d.Type),
		Value:       val.Float64,
		Description: desc,
		ValidFrom:   d.ValidFrom,
		ValidTo:     d.ValidTo,
		CreatedAt:   d.CreatedAt,
		ProductID:   d.ProductID,
		CategoryID:  d.CategoryID,
	}
}

func mapRole(r merchantstore.Role) Role {
	var desc *string
	if r.Description.Valid {
		desc = &r.Description.String
	}
	return Role{
		ID:          r.ID,
		MerchantID:  r.MerchantID,
		RoleType:    RoleType(r.RoleType),
		Description: desc,
		CreatedAt:   r.CreatedAt,
	}
}
