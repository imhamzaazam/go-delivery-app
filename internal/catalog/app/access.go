package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"
	merchantstore "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

func (service *Service) resolveViewerActor(ctx context.Context, merchantID uuid.UUID, email string) (actorstore.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
	actor, err := service.store.GetActorProfileByMerchantAndEmail(ctx, actorstore.GetActorProfileByMerchantAndEmailParams{
		MerchantID: merchantID,
		Email:      email,
	})
	if err != nil {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, domainerr.MatchPostgresError(err)
	}

	return actor, nil
}

func (service *Service) requireMerchantViewAccess(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, targetMerchantID uuid.UUID) (actorstore.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
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

func (service *Service) requireMerchantManageAccess(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, targetMerchantID uuid.UUID) (actorstore.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
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
