package app

import (
	"context"

	"github.com/google/uuid"

	commerce "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	coverage "github.com/horiondreher/go-web-api-boilerplate/internal/coverage"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) CreateMerchantServiceZoneByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, zoneID string, branchID string) (coverage.MerchantServiceZone, *domainerr.DomainError) {
	parsedMerchantID, merchantErr := utils.ParseUUID(merchantID, "merchant id")
	if merchantErr != nil {
		return coverage.MerchantServiceZone{}, merchantErr
	}
	parsedZoneID, zoneErr := utils.ParseUUID(zoneID, "zone id")
	if zoneErr != nil {
		return coverage.MerchantServiceZone{}, zoneErr
	}
	parsedBranchID, branchErr := utils.ParseUUID(branchID, "branch id")
	if branchErr != nil {
		return coverage.MerchantServiceZone{}, branchErr
	}

	if _, accessErr := service.requireMerchantManageAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return coverage.MerchantServiceZone{}, accessErr
	}

	if _, err := service.store.GetZone(ctx, parsedZoneID); err != nil {
		return coverage.MerchantServiceZone{}, domainerr.MatchPostgresError(err)
	}
	if _, err := service.store.GetBranch(ctx, commerce.GetBranchParams{MerchantID: parsedMerchantID, ID: parsedBranchID}); err != nil {
		return coverage.MerchantServiceZone{}, domainerr.MatchPostgresError(err)
	}

	serviceZone, err := service.store.CreateMerchantServiceZone(ctx, commerce.CreateMerchantServiceZoneParams{
		MerchantID: parsedMerchantID,
		ZoneID:     parsedZoneID,
		BranchID:   parsedBranchID,
	})
	if err != nil {
		return coverage.MerchantServiceZone{}, domainerr.MatchPostgresError(err)
	}

	return serviceZone, nil
}
