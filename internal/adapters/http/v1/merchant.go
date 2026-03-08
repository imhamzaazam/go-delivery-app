package v1

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	openapi_types "github.com/oapi-codegen/runtime/types"
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

func (adapter *HTTPAdapter) createMerchant(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	requestBody, err := httputils.Decode[CreateMerchantRequest](r)
	if err != nil {
		return err
	}

	validationErr := validate.Struct(createMerchantValidation{
		Name:          requestBody.Name,
		Ntn:           requestBody.Ntn,
		Address:       requestBody.Address,
		Category:      string(requestBody.Category),
		ContactNumber: requestBody.ContactNumber,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	createdMerchant, serviceErr := adapter.merchantService.CreateMerchant(r.Context(), ports.NewMerchant{
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

	err = httputils.Encode(w, r, http.StatusCreated, MerchantResponse{
		Id:            &responseID,
		Name:          &responseName,
		Ntn:           &responseNtn,
		Address:       &responseAddress,
		Category:      &responseCategory,
		ContactNumber: &responseContactNumber,
	})
	if err != nil {
		return err
	}

	return nil
}

func merchantResponse(merchantID openapi_types.UUID, name string, ntn string, address string, category string, contactNumber string, createdAt *time.Time, updatedAt *time.Time, logo *string) MerchantResponse {
	return MerchantResponse{
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

func (adapter *HTTPAdapter) CreateMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.handlerWrapper(adapter.createMerchant)(w, r)
}

func (adapter *HTTPAdapter) UpdateMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[UpdateMerchantRequest](r)
		if err != nil {
			return err
		}

		validationErr := validate.Struct(createMerchantValidation{
			Name:          requestBody.Name,
			Ntn:           requestBody.Ntn,
			Address:       requestBody.Address,
			Category:      string(requestBody.Category),
			ContactNumber: requestBody.ContactNumber,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		merchant, updateErr := adapter.merchantService.UpdateMerchant(r.Context(), authUser.MerchantID.String(), ports.NewMerchant{
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

		return httputils.Encode(w, r, http.StatusOK, merchantResponse(
			openapi_types.UUID(merchant.ID),
			merchant.Name,
			merchant.Ntn,
			merchant.Address,
			string(merchant.Category),
			merchant.ContactNumber,
			&createdAt,
			&updatedAt,
			nil,
		))
	})
}

func (adapter *HTTPAdapter) BootstrapMerchantActor(w http.ResponseWriter, r *http.Request, merchantID string) {
	adapter.handlerWrapper(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[BootstrapActorRequest](r)
		if err != nil {
			return err
		}

		validationErr := validate.Struct(bootstrapActorValidation{
			FullName: requestBody.FullName,
			Email:    string(requestBody.Email),
			Password: requestBody.Password,
			Role:     string(requestBody.Role),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		createdActor, serviceErr := adapter.merchantService.BootstrapActor(r.Context(), merchantID, ports.BootstrapActor{
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

		response := actorProfileResponse(
			openapi_types.UUID(parsedMerchantID),
			openapi_types.UUID(createdActor.UID),
			createdActor.FullName,
			createdActor.Email,
			true,
		)

		return httputils.Encode(w, r, http.StatusCreated, response)
	})(w, r)
}

func (adapter *HTTPAdapter) ListMerchants(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		merchant, err := adapter.merchantService.GetMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		createdAt := merchant.CreatedAt
		updatedAt := merchant.UpdatedAt
		response := []MerchantResponse{
			merchantResponse(
				openapi_types.UUID(merchant.ID),
				merchant.Name,
				merchant.Ntn,
				merchant.Address,
				string(merchant.Category),
				merchant.ContactNumber,
				&createdAt,
				&updatedAt,
				nil,
			),
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) GetMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		merchant, err := adapter.merchantService.GetMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}
		createdAt := merchant.CreatedAt
		updatedAt := merchant.UpdatedAt

		return httputils.Encode(w, r, http.StatusOK, merchantResponse(
			openapi_types.UUID(merchant.ID),
			merchant.Name,
			merchant.Ntn,
			merchant.Address,
			string(merchant.Category),
			merchant.ContactNumber,
			&createdAt,
			&updatedAt,
			nil,
		))
	})
}
