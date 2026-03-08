package v1

import (
	"fmt"
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type createActorValidation struct {
	FullName string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"omitempty"`
	Role     string `validate:"omitempty,oneof=merchant employee customer"`
}

func actorProfileResponse(merchantID openapi_types.UUID, uid openapi_types.UUID, fullName string, email string, isActive bool) ActorProfileResponse {
	responseEmail := openapi_types.Email(email)
	return ActorProfileResponse{
		MerchantId: &merchantID,
		Uid:        &uid,
		FullName:   &fullName,
		Email:      &responseEmail,
		IsActive:   &isActive,
	}
}

func (adapter *HTTPAdapter) createActor(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	requestBody, err := httputils.Decode[CreateActorRequest](r)
	if err != nil {
		return err
	}

	role := ""
	if requestBody.Role != nil {
		role = string(*requestBody.Role)
	}

	password := ""
	if requestBody.Password != nil {
		password = *requestBody.Password
	}

	validationErr := validate.Struct(createActorValidation{
		FullName: requestBody.FullName,
		Email:    string(requestBody.Email),
		Password: password,
		Role:     role,
	})
	if validationErr != nil {
		return httperr.MatchValidationError(validationErr)
	}

	isGuestCustomer := role == "customer" && password == ""
	if !isGuestCustomer && password == "" {
		return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "password is required for non-customer roles", fmt.Errorf("password is required for non-customer roles"))
	}

	authUser, authErr := adapter.currentAuthUser(r)
	if authErr != nil {
		return authErr
	}

	newActor := ports.NewActor{
		MerchantID: authUser.MerchantID,
		FullName:   requestBody.FullName,
		Email:      string(requestBody.Email),
		Password:   password,
	}

	var createdActor pgsqlc.CreateActorRow
	var createErr *domainerr.DomainError
	if requestBody.Role != nil {
		createdActor, createErr = adapter.commerceService.CreateActorByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), newActor, role)
	} else {
		createdActor, createErr = adapter.actorService.CreateActor(r.Context(), newActor)
	}
	if createErr != nil {
		return createErr
	}

	response := actorProfileResponse(
		authUser.MerchantID,
		createdActor.UID,
		createdActor.FullName,
		createdActor.Email,
		true,
	)

	return httputils.Encode(w, r, http.StatusCreated, response)
}

func (adapter *HTTPAdapter) getActorByUID(w http.ResponseWriter, r *http.Request, uid string) *domainerr.DomainError {
	authUser, authErr := adapter.currentAuthUser(r)
	if authErr != nil {
		return authErr
	}

	actor, err := adapter.actorService.GetActorByMerchantAndUID(r.Context(), authUser.MerchantID, uid)
	if err != nil {
		return err
	}

	response := actorProfileResponse(
		actor.MerchantID,
		actor.UID,
		actor.FullName,
		actor.Email,
		actor.IsActive,
	)
	if !actor.LastLogin.IsZero() {
		lastLogin := actor.LastLogin
		response.LastLogin = &lastLogin
	}

	return httputils.Encode(w, r, http.StatusOK, response)
}

func (adapter *HTTPAdapter) CreateActor(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, adapter.createActor)
}

func (adapter *HTTPAdapter) GetActorByUIDLegacy(w http.ResponseWriter, r *http.Request, uid string) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		return adapter.getActorByUID(w, r, uid)
	})
}

func (adapter *HTTPAdapter) GetAuthenticatedActor(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		actor, err := adapter.actorService.GetActorProfileByMerchantAndEmail(r.Context(), authUser.MerchantID, authUser.Email)
		if err != nil {
			return err
		}

		response := actorProfileResponse(
			actor.MerchantID,
			actor.UID,
			actor.FullName,
			actor.Email,
			actor.IsActive,
		)
		if !actor.LastLogin.IsZero() {
			lastLogin := actor.LastLogin
			response.LastLogin = &lastLogin
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) GetActorByUID(w http.ResponseWriter, r *http.Request, uid string) {
	adapter.GetActorByUIDLegacy(w, r, uid)
}

func (adapter *HTTPAdapter) ListActorsByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		actors, err := adapter.actorService.ListActorsByMerchant(r.Context(), authUser.MerchantID)
		if err != nil {
			return err
		}

		response := make([]ActorProfileResponse, 0, len(actors))
		for _, actor := range actors {
			item := actorProfileResponse(actor.MerchantID, actor.UID, actor.FullName, actor.Email, actor.IsActive)
			if !actor.LastLogin.IsZero() {
				lastLogin := actor.LastLogin
				item.LastLogin = &lastLogin
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) ListEmployeesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		employees, err := adapter.actorService.ListEmployeesByMerchant(r.Context(), authUser.MerchantID)
		if err != nil {
			return err
		}

		response := make([]ActorProfileResponse, 0, len(employees))
		for _, actor := range employees {
			item := actorProfileResponse(actor.MerchantID, actor.UID, actor.FullName, actor.Email, actor.IsActive)
			if !actor.LastLogin.IsZero() {
				lastLogin := actor.LastLogin
				item.LastLogin = &lastLogin
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}
