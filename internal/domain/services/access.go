package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (service *CommerceManager) resolveViewerActor(ctx context.Context, merchantID uuid.UUID, email string) (pgsqlc.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
	actor, err := service.store.GetActorProfileByMerchantAndEmail(ctx, pgsqlc.GetActorProfileByMerchantAndEmailParams{
		MerchantID: merchantID,
		Email:      email,
	})
	if err != nil {
		return pgsqlc.GetActorProfileByMerchantAndEmailRow{}, domainerr.MatchPostgresError(err)
	}

	return actor, nil
}

func (service *CommerceManager) requireMerchantViewAccess(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, targetMerchantID uuid.UUID) (pgsqlc.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
	viewer, viewerErr := service.resolveViewerActor(ctx, viewerMerchantID, viewerEmail)
	if viewerErr != nil {
		return pgsqlc.GetActorProfileByMerchantAndEmailRow{}, viewerErr
	}

	canView, allowErr := service.canViewMerchant(ctx, viewer.UID, targetMerchantID)
	if allowErr != nil {
		return pgsqlc.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewInternalError(allowErr)
	}
	if !canView {
		return pgsqlc.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "not allowed to view merchant", fmt.Errorf("not allowed to view merchant"))
	}

	return viewer, nil
}

func (service *CommerceManager) requireMerchantManageAccess(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, targetMerchantID uuid.UUID) (pgsqlc.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
	viewer, viewerErr := service.resolveViewerActor(ctx, viewerMerchantID, viewerEmail)
	if viewerErr != nil {
		return pgsqlc.GetActorProfileByMerchantAndEmailRow{}, viewerErr
	}

	isAdmin, adminErr := service.hasRole(ctx, viewer.UID, pgsqlc.RoleTypeAdmin, uuid.Nil)
	if adminErr != nil {
		return pgsqlc.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewInternalError(adminErr)
	}
	if isAdmin {
		return viewer, nil
	}

	isMerchant, merchantErr := service.hasRole(ctx, viewer.UID, pgsqlc.RoleTypeMerchant, targetMerchantID)
	if merchantErr != nil {
		return pgsqlc.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewInternalError(merchantErr)
	}
	if !isMerchant {
		return pgsqlc.GetActorProfileByMerchantAndEmailRow{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "merchant role required", fmt.Errorf("merchant role required"))
	}

	return viewer, nil
}
