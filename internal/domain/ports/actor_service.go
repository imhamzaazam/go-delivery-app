package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type NewActor struct {
	MerchantID uuid.UUID
	FullName   string
	Email      string
	Password   string
}

type ActorService interface {
	CreateActor(ctx context.Context, newActor NewActor) (pgsqlc.CreateActorRow, *domainerr.DomainError)
	LoginActor(ctx context.Context, loginActor LoginActor) (pgsqlc.GetActorRow, *domainerr.DomainError)
	CreateActorSession(ctx context.Context, newActorSession NewActorSession) (pgsqlc.CreateSessionRow, *domainerr.DomainError)
	GetActorSession(ctx context.Context, refreshTokenID uuid.UUID) (pgsqlc.GetSessionRow, *domainerr.DomainError)
	GetActorByUID(ctx context.Context, actorUID string) (pgsqlc.GetActorByUIDRow, *domainerr.DomainError)
	GetActorByMerchantAndUID(ctx context.Context, merchantID uuid.UUID, actorUID string) (pgsqlc.GetActorByUIDRow, *domainerr.DomainError)
	GetActorProfileByMerchantAndEmail(ctx context.Context, merchantID uuid.UUID, email string) (pgsqlc.GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError)
	ListActorsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListActorsByMerchantRow, *domainerr.DomainError)
	ListEmployeesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]pgsqlc.ListEmployeesByMerchantRow, *domainerr.DomainError)
}
