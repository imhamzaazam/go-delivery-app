package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (service *ActorManager) GetActorProfileByMerchantAndEmail(ctx context.Context, merchantID uuid.UUID, email string) (pgsqlc.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError) {
	actor, err := service.store.GetActorProfileByMerchantAndEmail(ctx, pgsqlc.GetActorProfileByMerchantAndEmailParams{
		MerchantID: merchantID,
		Email:      email,
	})
	if err != nil {
		return pgsqlc.GetActorProfileByMerchantAndEmailRow{}, domainerr.MatchPostgresError(err)
	}

	return actor, nil
}

func (service *ActorManager) ListActorsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListActorsByMerchantRow, *domainerr.DomainError) {
	actors, err := service.store.ListActorsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return actors, nil
}

func (service *ActorManager) ListEmployeesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListEmployeesByMerchantRow, *domainerr.DomainError) {
	employees, err := service.store.ListEmployeesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	return employees, nil
}
