package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (adapter *HTTPAdapter) currentActorProfile(r *http.Request) (*actorProfileView, *domainerr.DomainError) {
	authUser, authErr := adapter.currentAuthUser(r)
	if authErr != nil {
		return nil, authErr
	}

	actor, err := adapter.actorService.GetActorProfileByMerchantAndEmail(r.Context(), authUser.MerchantID, authUser.Email)
	if err != nil {
		return nil, err
	}

	return &actorProfileView{
		MerchantID: actor.MerchantID,
		UID:        actor.UID,
		Email:      actor.Email,
	}, nil
}

type actorProfileView struct {
	MerchantID uuid.UUID
	UID        uuid.UUID
	Email      string
}
