package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

func (service *Service) GetActorProfileByMerchantAndEmail(ctx context.Context, merchantID uuid.UUID, email string) (actorstore.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
	actor, err := service.store.GetActorProfileByMerchantAndEmail(ctx, actorstore.GetActorProfileByMerchantAndEmailParams{
		MerchantID: merchantID,
		Email:      email,
	})
	if err != nil {
		return actorstore.GetActorProfileByMerchantAndEmailRow{}, domainerr.MatchPostgresError(err)
	}

	return actor, nil
}

func (service *Service) ListActorsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]actorstore.ListActorsByMerchantRow, *domainerr.DomainError) {
	actors, err := service.store.ListActorsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return actors, nil
}

func (service *Service) ListEmployeesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]actorstore.ListEmployeesByMerchantRow, *domainerr.DomainError) {
	employees, err := service.store.ListEmployeesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return employees, nil
}
