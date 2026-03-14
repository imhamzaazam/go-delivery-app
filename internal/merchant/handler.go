package merchant

import (
	"errors"
	"fmt"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
	middleware "github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/middleware"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/token"
)

type Handler struct {
	service    *MerchantService
	config     *utils.Config
	tokenMaker *token.JWTMaker
	validate   *validator.Validate
}

type domainHandler func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError

type authUser struct {
	Email      string
	MerchantID uuid.UUID
}

const (
	gatewayMerchantIDHeader = "X-Merchant-Id"
	gatewaySecretHeader     = "X-Gateway-Secret"
)

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

func New(service *MerchantService, config *utils.Config, tokenMaker *token.JWTMaker, validate *validator.Validate) *Handler {
	return &Handler{
		service:    service,
		config:     config,
		tokenMaker: tokenMaker,
		validate:   validate,
	}
}

func (handler *Handler) wrap(handlerFn domainHandler) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if apiErr := handlerFn(w, r); apiErr != nil {
			var httpErrIntf *domainerr.DomainError
			var err *domainerr.DomainError

			requestID := middleware.GetRequestID(r.Context())

			if errors.As(apiErr, &httpErrIntf) {
				log.Info().Str("id", requestID).Str("error message", httpErrIntf.OriginalError).Msg("request error")
				err = httputils.Encode(w, r, httpErrIntf.HTTPCode, httpErrIntf.HTTPErrorBody)
			} else {
				stdhttp.Error(w, "Internal server error", stdhttp.StatusInternalServerError)
			}

			if err != nil {
				log.Err(err).Msg("error encoding response")
			}
		}
	}
}

func (handler *Handler) serveAuthenticated(w stdhttp.ResponseWriter, r *stdhttp.Request, next domainHandler) {
	authenticatedHandler := middleware.Authentication(handler.tokenMaker)(handler.wrap(next))
	authenticatedHandler.ServeHTTP(w, r)
}

func (handler *Handler) currentAuthUser(r *stdhttp.Request) (*authUser, *domainerr.DomainError) {
	auth := middleware.GetAuthUser(r.Context())
	if auth == nil {
		return nil, domainerr.NewDomainError(stdhttp.StatusUnauthorized, domainerr.UnauthorizedError, "unauthorized", errors.New("unauthorized"))
	}

	return &authUser{Email: auth.Email, MerchantID: auth.MerchantID}, nil
}

func (handler *Handler) currentMerchantID(r *stdhttp.Request) (uuid.UUID, *domainerr.DomainError) {
	if strings.TrimSpace(r.Header.Get("Authorization")) != "" {
		authUser, authErr := handler.authUserFromAuthorizationHeader(r)
		if authErr != nil {
			return uuid.Nil, authErr
		}

		return authUser.MerchantID, nil
	}

	return handler.trustedGatewayMerchantID(r)
}

func (handler *Handler) authUserFromAuthorizationHeader(r *stdhttp.Request) (*authUser, *domainerr.DomainError) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	fields := strings.Fields(auth)
	if len(fields) < 2 {
		return nil, domainerr.NewDomainError(stdhttp.StatusUnauthorized, domainerr.UnauthorizedError, "invalid authorization header", errors.New("invalid authorization header"))
	}
	if strings.ToLower(fields[0]) != "bearer" {
		return nil, domainerr.NewDomainError(stdhttp.StatusUnauthorized, domainerr.UnauthorizedError, "invalid authorization type", errors.New("invalid authorization type"))
	}

	payload, err := handler.tokenMaker.VerifyToken(fields[1])
	if err != nil {
		return nil, err
	}

	return &authUser{Email: payload.Email, MerchantID: payload.MerchantID}, nil
}

func (handler *Handler) trustedGatewayMerchantID(r *stdhttp.Request) (uuid.UUID, *domainerr.DomainError) {
	if handler.config == nil || strings.TrimSpace(handler.config.GatewaySharedSecret) == "" {
		return uuid.Nil, domainerr.NewDomainError(stdhttp.StatusUnauthorized, domainerr.UnauthorizedError, "missing trusted merchant context", errors.New("gateway shared secret is not configured"))
	}

	providedSecret := strings.TrimSpace(r.Header.Get(gatewaySecretHeader))
	if providedSecret == "" || providedSecret != handler.config.GatewaySharedSecret {
		return uuid.Nil, domainerr.NewDomainError(stdhttp.StatusUnauthorized, domainerr.UnauthorizedError, "missing trusted merchant context", errors.New("invalid gateway secret"))
	}

	merchantHeader := strings.TrimSpace(r.Header.Get(gatewayMerchantIDHeader))
	if merchantHeader == "" {
		return uuid.Nil, domainerr.NewDomainError(stdhttp.StatusUnauthorized, domainerr.UnauthorizedError, "missing trusted merchant context", errors.New("missing merchant header"))
	}

	merchantID, parseErr := uuid.Parse(merchantHeader)
	if parseErr != nil {
		return uuid.Nil, domainerr.NewDomainError(stdhttp.StatusUnauthorized, domainerr.UnauthorizedError, "invalid trusted merchant context", parseErr)
	}

	if _, err := handler.service.GetMerchant(r.Context(), merchantID.String()); err != nil {
		return uuid.Nil, domainerr.NewDomainError(stdhttp.StatusUnauthorized, domainerr.UnauthorizedError, "invalid trusted merchant context", err)
	}

	return merchantID, nil
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

func merchantResponse(merchantID uuid.UUID, name string, ntn string, address string, category string, contactNumber string, createdAt *time.Time, updatedAt *time.Time, logo *string) api.MerchantResponse {
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

func branchResponse(branch Branch) api.BranchResponse {
	city := string(branch.City)
	openingTime := utils.FormatBranchClock(branch.OpeningTimeMinutes)
	closingTime := utils.FormatBranchClock(branch.ClosingTimeMinutes)
	isOpen := utils.IsBranchOpenAt(branch.OpeningTimeMinutes, branch.ClosingTimeMinutes, time.Now())
	createdAt := branch.CreatedAt
	updatedAt := branch.UpdatedAt
	return api.BranchResponse{
		Id:            utils.PtrUUID(branch.ID),
		MerchantId:    utils.PtrUUID(branch.MerchantID),
		Name:          &branch.Name,
		Address:       &branch.Address,
		ContactNumber: branch.ContactNumber,
		City:          &city,
		OpeningTime:   &openingTime,
		ClosingTime:   &closingTime,
		IsOpen:        &isOpen,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}
}

func branchAvailabilityResponse(availability BranchAvailability) api.BranchAvailabilityResponse {
	branchID := availability.Branch.ID
	merchantID := availability.Branch.MerchantID
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

func discountResponse(discount MerchantDiscount) api.DiscountResponse {
	discountType := string(discount.Type)
	value := discount.Value
	createdAt := discount.CreatedAt
	response := api.DiscountResponse{
		Id:          utils.PtrUUID(discount.ID),
		MerchantId:  utils.PtrUUID(discount.MerchantID),
		Type:        &discountType,
		Value:       &value,
		Description: discount.Description,
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
		productID := discount.ProductID
		response.ProductId = &productID
	}
	if discount.CategoryID != uuid.Nil {
		categoryID := discount.CategoryID
		response.CategoryId = &categoryID
	}
	return response
}

func (handler *Handler) createMerchant(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
	requestBody, err := httputils.Decode[api.CreateMerchantRequest](r)
	if err != nil {
		return err
	}

	validationErr := handler.validate.Struct(createMerchantValidation{
		Name:          requestBody.Name,
		Ntn:           requestBody.Ntn,
		Address:       requestBody.Address,
		Category:      string(requestBody.Category),
		ContactNumber: requestBody.ContactNumber,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	createdMerchant, serviceErr := handler.service.CreateMerchant(r.Context(), NewMerchant{
		Name:          requestBody.Name,
		Ntn:           requestBody.Ntn,
		Address:       requestBody.Address,
		Category:      string(requestBody.Category),
		ContactNumber: requestBody.ContactNumber,
	})
	if serviceErr != nil {
		return serviceErr
	}

	responseID := createdMerchant.ID
	responseName := createdMerchant.Name
	responseNtn := createdMerchant.Ntn
	responseAddress := createdMerchant.Address
	responseCategory := string(createdMerchant.Category)
	responseContactNumber := createdMerchant.ContactNumber

	return httputils.Encode(w, r, stdhttp.StatusCreated, api.MerchantResponse{
		Id:            &responseID,
		Name:          &responseName,
		Ntn:           &responseNtn,
		Address:       &responseAddress,
		Category:      &responseCategory,
		ContactNumber: &responseContactNumber,
	})
}

func (handler *Handler) CreateMerchant(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handler.wrap(handler.createMerchant)(w, r)
}

func (handler *Handler) UpdateMerchant(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handler.serveAuthenticated(w, r, func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.UpdateMerchantRequest](r)
		if err != nil {
			return err
		}

		validationErr := handler.validate.Struct(createMerchantValidation{
			Name:          requestBody.Name,
			Ntn:           requestBody.Ntn,
			Address:       requestBody.Address,
			Category:      string(requestBody.Category),
			ContactNumber: requestBody.ContactNumber,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := handler.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		updatedMerchant, updateErr := handler.service.UpdateMerchant(r.Context(), authUser.MerchantID.String(), NewMerchant{
			Name:          requestBody.Name,
			Ntn:           requestBody.Ntn,
			Address:       requestBody.Address,
			Category:      string(requestBody.Category),
			ContactNumber: requestBody.ContactNumber,
		})
		if updateErr != nil {
			return updateErr
		}

		createdAt := updatedMerchant.CreatedAt
		updatedAt := updatedMerchant.UpdatedAt
		return httputils.Encode(w, r, stdhttp.StatusOK, merchantResponse(updatedMerchant.ID, updatedMerchant.Name, updatedMerchant.Ntn, updatedMerchant.Address, string(updatedMerchant.Category), updatedMerchant.ContactNumber, &createdAt, &updatedAt, nil))
	})
}

func (handler *Handler) BootstrapMerchantActor(w stdhttp.ResponseWriter, r *stdhttp.Request, merchantID string) {
	handler.wrap(func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.BootstrapActorRequest](r)
		if err != nil {
			return err
		}

		validationErr := handler.validate.Struct(bootstrapActorValidation{
			FullName: requestBody.FullName,
			Email:    string(requestBody.Email),
			Password: requestBody.Password,
			Role:     string(requestBody.Role),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		createdActor, serviceErr := handler.service.BootstrapActor(r.Context(), merchantID, BootstrapActor{
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
			return domainerr.NewDomainError(stdhttp.StatusBadRequest, domainerr.ValidationError, "invalid merchant id", parseErr)
		}

		response := actorProfileResponse(parsedMerchantID, createdActor.UID, createdActor.FullName, createdActor.Email, true)
		return httputils.Encode(w, r, stdhttp.StatusCreated, response)
	})(w, r)
}

func (handler *Handler) ListMerchants(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handler.serveAuthenticated(w, r, func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
		authUser, authErr := handler.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		merchantRow, err := handler.service.GetMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		createdAt := merchantRow.CreatedAt
		updatedAt := merchantRow.UpdatedAt
		response := []api.MerchantResponse{
			merchantResponse(merchantRow.ID, merchantRow.Name, merchantRow.Ntn, merchantRow.Address, string(merchantRow.Category), merchantRow.ContactNumber, &createdAt, &updatedAt, nil),
		}

		return httputils.Encode(w, r, stdhttp.StatusOK, response)
	})
}

func (handler *Handler) GetMerchant(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handler.serveAuthenticated(w, r, func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
		authUser, authErr := handler.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		merchantRow, err := handler.service.GetMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}
		createdAt := merchantRow.CreatedAt
		updatedAt := merchantRow.UpdatedAt

		return httputils.Encode(w, r, stdhttp.StatusOK, merchantResponse(merchantRow.ID, merchantRow.Name, merchantRow.Ntn, merchantRow.Address, string(merchantRow.Category), merchantRow.ContactNumber, &createdAt, &updatedAt, nil))
	})
}

func (handler *Handler) ListBranchesByMerchant(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handler.serveAuthenticated(w, r, func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
		authUser, authErr := handler.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		branches, err := handler.service.ListBranchesByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.BranchResponse, 0, len(branches))
		for _, branch := range branches {
			response = append(response, branchResponse(branch))
		}

		return httputils.Encode(w, r, stdhttp.StatusOK, response)
	})
}

func (handler *Handler) CreateBranchByMerchant(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handler.serveAuthenticated(w, r, func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateBranchRequest](r)
		if err != nil {
			return err
		}
		validationErr := handler.validate.Struct(createBranchValidation{
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

		authUser, authErr := handler.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		branch, createErr := handler.service.CreateBranchByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), requestBody.Name, requestBody.Address, requestBody.ContactNumber, string(requestBody.City), requestBody.OpeningTime, requestBody.ClosingTime)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, stdhttp.StatusCreated, branchResponse(branch))
	})
}

func (handler *Handler) GetBranchAvailability(w stdhttp.ResponseWriter, r *stdhttp.Request, branchID string) {
	handler.wrap(func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
		merchantID, merchantErr := handler.currentMerchantID(r)
		if merchantErr != nil {
			return merchantErr
		}

		availability, err := handler.service.GetBranchAvailability(r.Context(), merchantID.String(), branchID)
		if err != nil {
			return err
		}

		return httputils.Encode(w, r, stdhttp.StatusOK, branchAvailabilityResponse(availability))
	})(w, r)
}

func (handler *Handler) ListDiscountsByMerchant(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handler.serveAuthenticated(w, r, func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
		authUser, authErr := handler.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		discounts, err := handler.service.ListDiscountsByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.DiscountResponse, 0, len(discounts))
		for _, discount := range discounts {
			response = append(response, discountResponse(discount))
		}

		return httputils.Encode(w, r, stdhttp.StatusOK, response)
	})
}

func (handler *Handler) CreateDiscountByMerchant(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handler.serveAuthenticated(w, r, func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
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
		validationErr := handler.validate.Struct(createDiscountValidation{Type: string(requestBody.Type), Value: requestBody.Value, ProductID: productID, CategoryID: categoryID})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}
		if productID != "" && categoryID != "" {
			return domainerr.NewDomainError(stdhttp.StatusBadRequest, domainerr.ValidationError, "discount scope must target either product or category", fmt.Errorf("discount scope must target either product or category"))
		}

		authUser, authErr := handler.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		description := ""
		if requestBody.Description != nil {
			description = *requestBody.Description
		}
		discount, createErr := handler.service.CreateMerchantDiscountByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), string(requestBody.Type), requestBody.Value, description, productID, categoryID)
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, stdhttp.StatusCreated, discountResponse(discount))
	})
}

func (handler *Handler) ListRolesByMerchant(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handler.serveAuthenticated(w, r, func(w stdhttp.ResponseWriter, r *stdhttp.Request) *domainerr.DomainError {
		authUser, authErr := handler.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		roles, err := handler.service.ListRolesByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]api.RoleResponse, 0, len(roles))
		for _, role := range roles {
			roleType := string(role.RoleType)
			createdAt := role.CreatedAt
			response = append(response, api.RoleResponse{
				Id:          utils.PtrUUID(role.ID),
				MerchantId:  utils.PtrUUID(role.MerchantID),
				RoleType:    &roleType,
				Description: role.Description,
				CreatedAt:   &createdAt,
			})
		}

		return httputils.Encode(w, r, stdhttp.StatusOK, response)
	})
}
