package presentation

import (
	"fmt"
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	actor "github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

type Handler struct {
	shared *core.Shared
}

type createActorValidation struct {
	FullName string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"omitempty"`
	Role     string `validate:"omitempty,oneof=merchant employee customer"`
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

func (handler *Handler) createActor(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
	requestBody, err := httputils.Decode[api.CreateActorRequest](r)
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

	validationErr := handler.shared.Validate.Struct(createActorValidation{
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

	authUser, authErr := handler.shared.CurrentAuthUser(r)
	if authErr != nil {
		return authErr
	}

	newActor := actor.NewActor{
		MerchantID: authUser.MerchantID,
		FullName:   requestBody.FullName,
		Email:      string(requestBody.Email),
		Password:   password,
	}

	var createdActor actorstore.CreateActorRow
	var createErr *domainerr.DomainError
	if requestBody.Role != nil {
		createdActor, createErr = handler.shared.CommerceService.CreateActorByMerchant(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), newActor, role)
	} else {
		createdActor, createErr = handler.shared.ActorService.CreateActor(r.Context(), newActor)
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

func (handler *Handler) getActorByUID(w http.ResponseWriter, r *http.Request, uid string) *domainerr.DomainError {
	authUser, authErr := handler.shared.CurrentAuthUser(r)
	if authErr != nil {
		return authErr
	}

	actor, err := handler.shared.ActorService.GetActorByMerchantAndUID(r.Context(), authUser.MerchantID, uid)
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

func (handler *Handler) CreateActor(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, handler.createActor)
}

func (handler *Handler) GetActorByUIDLegacy(w http.ResponseWriter, r *http.Request, uid string) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		return handler.getActorByUID(w, r, uid)
	})
}

func (handler *Handler) GetAuthenticatedActor(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		actor, err := handler.shared.ActorService.GetActorProfileByMerchantAndEmail(r.Context(), authUser.MerchantID, authUser.Email)
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

func (handler *Handler) GetActorByUID(w http.ResponseWriter, r *http.Request, uid string) {
	handler.GetActorByUIDLegacy(w, r, uid)
}

func (handler *Handler) ListActorsByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		actors, err := handler.shared.ActorService.ListActorsByMerchant(r.Context(), authUser.MerchantID)
		if err != nil {
			return err
		}

		response := make([]api.ActorProfileResponse, 0, len(actors))
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

func (handler *Handler) ListEmployeesByMerchant(w http.ResponseWriter, r *http.Request) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		employees, err := handler.shared.ActorService.ListEmployeesByMerchant(r.Context(), authUser.MerchantID)
		if err != nil {
			return err
		}

		response := make([]api.ActorProfileResponse, 0, len(employees))
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
