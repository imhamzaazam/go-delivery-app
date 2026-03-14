package app

import (
	"context"

	"github.com/google/uuid"
	commerce "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
)

func (service *Service) canViewMerchant(ctx context.Context, actorID uuid.UUID, merchantID uuid.UUID) (bool, error) {
	isAdmin, adminErr := service.hasRole(ctx, actorID, commerce.RoleTypeAdmin, uuid.Nil)
	if adminErr != nil {
		return false, adminErr
	}
	if isAdmin {
		return true, nil
	}

	isMerchant, merchantErr := service.hasRole(ctx, actorID, commerce.RoleTypeMerchant, merchantID)
	if merchantErr != nil {
		return false, merchantErr
	}
	return isMerchant, nil
}
