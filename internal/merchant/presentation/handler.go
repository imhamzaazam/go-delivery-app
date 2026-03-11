package presentation

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	merchant "github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

type Handler struct {
	shared *core.Shared
}

type createMerchantValidation struct {
	Name          string `validate:"required"`
	Ntn           string `validate:"required"`
	Address       string `validate:"required"`
	Category      string `validate:"required,oneof=restaurant pharma bakery"`
	ContactNumber string `validate:"required"`
}

type bootstrapActorValidation struct {
	FullName string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
	Role     string `validate:"required,oneof=merchant employee customer"`
}

type createBranchValidation struct {
	Name          string `validate:"required"`
	Address       string `validate:"required"`
	ContactNumber string `validate:"required"`
	City          string `validate:"required,oneof=Karachi Lahore"`
	OpeningTime   string `validate:"required"`
	ClosingTime   string `validate:"required"`
}

type createDiscountValidation struct {
	Type       string  `validate:"required,oneof=flat percentage"`
	Value      float64 `validate:"required,gt=0"`
	ProductID  string  `validate:"omitempty,uuid"`
	CategoryID string  `validate:"omitempty,uuid"`
}

func New(shared *core.Shared) *Handler {
	return &Handler{shared: shared}
}

func actorProfileResponse(merchantID openapi_types.UUID, uid openapi_types.UUID, fullName string, email string, isActive bool) api.ActorProfileResponse {
	responseEmail := openapi_types.Email(email)
	return api.ActorProfileResponse{
		MerchantId: &merchantID,
		Uid:        &uid,
		FullName:   &fullName,
		Email:      &responseEmail,
		IsActive:   &isActive,
	}
}

func merchantResponse(merchantID openapi_types.UUID, name string, ntn string, address string, category string, contactNumber string, createdAt *time.Time, updatedAt *time.Time, logo *string) api.MerchantResponse {
	return api.MerchantResponse{
		Id:            &merchantID,
		Name:          &name,
		Ntn:           &ntn,
		Address:       &address,
		Category:      &category,
		ContactNumber: &contactNumber,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		Logo:          logo,
	}
}

func branchResponse(branch merchantstore.Branch) api.BranchResponse {
	contactNumber := core.TextString(branch.ContactNumber)
	city := string(branch.City)
	openingTime := utils.FormatBranchClock(branch.OpeningTimeMinutes)
	closingTime := utils.FormatBranchClock(branch.ClosingTimeMinutes)
	isOpen := utils.IsBranchOpenAt(branch.OpeningTimeMinutes, branch.ClosingTimeMinutes, time.Now())
	createdAt := branch.CreatedAt
	updatedAt := branch.UpdatedAt
	return api.BranchResponse{
		Id:            core.PtrUUID(branch.ID),
		MerchantId:    core.PtrUUID(branch.MerchantID),
		Name:          &branch.Name,
		Address:       &branch.Address,
		ContactNumber: &contactNumber,
		City:          &city,
		OpeningTime:   &openingTime,
		ClosingTime:   &closingTime,
		IsOpen:        &isOpen,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}
}

func branchAvailabilityResponse(availability merchant.BranchAvailability) api.BranchAvailabilityResponse {
	branchID := openapi_types.UUID(availability.Branch.ID)
	merchantID := openapi_types.UUID(availability.Branch.MerchantID)
	branchName := availability.Branch.Name
	return api.BranchAvailabilityResponse{
		MerchantId:  &merchantID,
		BranchId:    &branchID,
		BranchName:  &branchName,
		IsOpen:      &availability.IsOpen,
		OpeningTime: &availability.OpeningTime,
		ClosingTime: &availability.ClosingTime,
		CurrentTime: &availability.CurrentTime,
		Timezone:    &availability.TimezoneName,
	}
}

func discountResponse(discount merchantstore.MerchantDiscount) api.DiscountResponse {
	description := core.TextString(discount.Description)
	discountType := string(discount.Type)
	value := core.NumericToFloat64(discount.Value)
	createdAt := discount.CreatedAt
	response := api.DiscountResponse{
		Id:          core.PtrUUID(discount.ID),
		MerchantId:  core.PtrUUID(discount.MerchantID),
		Type:        &discountType,
		Value:       &value,
		Description: &description,
		CreatedAt:   &createdAt,
	}
	if !discount.ValidFrom.IsZero() {
		validFrom := discount.ValidFrom
		response.ValidFrom = &validFrom
	}
	if !discount.ValidTo.IsZero() {
		validTo := discount.ValidTo
		response.ValidTo = &validTo
	}
	if discount.ProductID != uuid.Nil {
		productID := openapi_types.UUID(discount.ProductID)
		response.ProductId = &productID
	}
	if discount.CategoryID != uuid.Nil {
		categoryID := openapi_types.UUID(discount.CategoryID)
		response.CategoryId = &categoryID
	}
	return response
}

func (handler *Handler) createMerchant(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	requestBody, err := httputils.Decode[api.CreateMerchantRequest](r)
	if err != nil {
		return err
	}

	validationErr := handler.shared.Validate.Struct(createMerchantValidation{
		Name:          requestBody.Name,
		Ntn:           requestBody.Ntn,
		Address:       requestBody.Address,
		Category:      string(requestBody.Category),
		ContactNumber: requestBody.ContactNumber,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	createdMerchant, serviceErr := handler.shared.MerchantService.CreateMerchant(r.Context(), merchant.NewMerchant{
		Name:          requestBody.Name,
		Ntn:           requestBody.Ntn,
		Address:       requestBody.Address,
		Category:      string(requestBody.Category),
		ContactNumber: requestBody.ContactNumber,
	})
	if serviceErr != nil {
		return serviceErr
	}

	responseID := openapi_types.UUID(createdMerchant.ID)
	responseName := createdMerchant.Name
	responseNtn := createdMerchant.Ntn
	responseAddress := createdMerchant.Address
	responseCategory := string(createdMerchant.Category)
	responseContactNumber := createdMerchant.ContactNumber

	return httputils.Encode(w, r, http.StatusCreated, api.MerchantResponse{
		Id:            &responseID,
		Name:          &responseName,
		Ntn:           &responseNtn,
		Address:       &responseAddress,
		Category:      &responseCategory,
		ContactNumber: &responseContactNumber,
	})
}

func (handler *Handler) CreateMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.Wrap(handler.createMerchant)(w, r)
}

func (handler *Handler) UpdateMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.UpdateMerchantRequest](r)
		if err != nil {
			return err
		}

		validationErr := handler.shared.Validate.Struct(createMerchantValidation{
			Name:          requestBody.Name,
			Ntn:           requestBody.Ntn,
			Address:       requestBody.Address,
			Category:      string(requestBody.Category),
			ContactNumber: requestBody.ContactNumber,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		merchant, updateErr := handler.shared.MerchantService.UpdateMerchant(r.Context(), authUser.MerchantID.String(), merchant.NewMerchant{
			Name:          requestBody.Name,
			Ntn:           requestBody.Ntn,
			Address:       requestBody.Address,
			Category:      string(requestBody.Category),
			ContactNumber: requestBody.ContactNumber,
		})
		if updateErr != nil {
			return updateErr
		}

		createdAt := merchant.CreatedAt
		updatedAt := merchant.UpdatedAt
		return httputils.Encode(w, r, http.StatusOK, merchantResponse(openapi_types.UUID(merchant.ID), merchant.Name, merchant.Ntn, merchant.Address, string(merchant.Category), merchant.ContactNumber, &createdAt, &updatedAt, nil))
	})
}

func (handler *Handler) BootstrapMerchantActor(w http.ResponseWriter, r *http.Request, merchantID string) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.BootstrapActorRequest](r)
		if err != nil {
			return err
		}

		validationErr := handler.shared.Validate.Struct(bootstrapActorValidation{
			FullName: requestBody.FullName,
			Email:    string(requestBody.Email),
			Password: requestBody.Password,
			Role:     string(requestBody.Role),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		createdActor, serviceErr := handler.shared.MerchantService.BootstrapActor(r.Context(), merchantID, merchant.BootstrapActor{
			FullName: requestBody.FullName,
			Email:    string(requestBody.Email),
			Password: requestBody.Password,
			Role:     string(requestBody.Role),
		})
		if serviceErr != nil {
			return serviceErr
		}

		parsedMerchantID, parseErr := uuid.Parse(merchantID)
		if parseErr != nil {
			return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid merchant id", parseErr)
		}

		response := actorProfileResponse(openapi_types.UUID(parsedMerchantID), openapi_types.UUID(createdActor.UID), createdActor.FullName, createdActor.Email, true)
		return httputils.Encode(w, r, http.StatusCreated, response)
	})(w, r)
}

func (handler *Handler) ListMerchants(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		merchant, err := handler.shared.MerchantService.GetMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		createdAt := merchant.CreatedAt
		updatedAt := merchant.UpdatedAt
		response := []api.MerchantResponse{
			merchantResponse(openapi_types.UUID(merchant.ID), merchant.Name, merchant.Ntn, merchant.Address, string(merchant.Category), merchant.ContactNumber, &createdAt, &updatedAt, nil),
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (handler *Handler) GetMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		merchant, err := handler.shared.MerchantService.GetMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}
		createdAt := merchant.CreatedAt
		updatedAt := merchant.UpdatedAt

		return httputils.Encode(w, r, http.StatusOK, merchantResponse(openapi_types.UUID(merchant.ID), merchant.Name, merchant.Ntn, merchant.Address, string(merchant.Category), merchant.ContactNumber, &createdAt, &updatedAt, nil))
	})
}

func (handler *Handler) ListBranchesByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		branches, err := handler.shared.MerchantService.ListBranchesByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.BranchResponse, 0, len(branches))
		for _, branch := range branches {
			response = append(response, branchResponse(branch))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (handler *Handler) CreateBranchByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateBranchRequest](r)
		if err != nil {
			return err
		}
		validationErr := handler.shared.Validate.Struct(createBranchValidation{
			Name:          requestBody.Name,
			Address:       requestBody.Address,
			ContactNumber: requestBody.ContactNumber,
			City:          string(requestBody.City),
			OpeningTime:   requestBody.OpeningTime,
			ClosingTime:   requestBody.ClosingTime,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		branch, createErr := handler.shared.CommerceService.CreateBranchByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.Name, requestBody.Address, requestBody.ContactNumber, string(requestBody.City), requestBody.OpeningTime, requestBody.ClosingTime)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, branchResponse(branch))
	})
}

func (handler *Handler) GetBranchAvailability(w http.ResponseWriter, r *http.Request, branchID string) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		merchantID, merchantErr := handler.shared.CurrentMerchantID(r)
		if merchantErr != nil {
			return merchantErr
		}

		availability, err := handler.shared.ReadService.GetBranchAvailability(r.Context(), merchantID.String(), branchID)
		if err != nil {
			return err
		}

		return httputils.Encode(w, r, http.StatusOK, branchAvailabilityResponse(availability))
	})(w, r)
}

func (handler *Handler) ListDiscountsByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		discounts, err := handler.shared.MerchantService.ListDiscountsByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.DiscountResponse, 0, len(discounts))
		for _, discount := range discounts {
			response = append(response, discountResponse(discount))
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (handler *Handler) CreateDiscountByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateDiscountRequest](r)
		if err != nil {
			return err
		}
		productID := ""
		if requestBody.ProductId != nil {
			productID = requestBody.ProductId.String()
		}
		categoryID := ""
		if requestBody.CategoryId != nil {
			categoryID = requestBody.CategoryId.String()
		}
		validationErr := handler.shared.Validate.Struct(createDiscountValidation{Type: string(requestBody.Type), Value: requestBody.Value, ProductID: productID, CategoryID: categoryID})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}
		if productID != "" && categoryID != "" {
			return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "discount scope must target either product or category", fmt.Errorf("discount scope must target either product or category"))
		}

		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		description := ""
		if requestBody.Description != nil {
			description = *requestBody.Description
		}
		discount, createErr := handler.shared.CommerceService.CreateMerchantDiscountByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), string(requestBody.Type), requestBody.Value, description, productID, categoryID)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, discountResponse(discount))
	})
}

func (handler *Handler) ListRolesByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		roles, err := handler.shared.MerchantService.ListRolesByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.RoleResponse, 0, len(roles))
		for _, role := range roles {
			description := role.Description.String
			roleType := string(role.RoleType)
			createdAt := role.CreatedAt
			response = append(response, api.RoleResponse{
				Id:          core.PtrUUID(role.ID),
				MerchantId:  core.PtrUUID(role.MerchantID),
				RoleType:    &roleType,
				Description: &description,
				CreatedAt:   &createdAt,
			})
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}
