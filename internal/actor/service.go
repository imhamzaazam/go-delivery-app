package actor

import (
	"context"
	"time"

	"github.com/google/uuid"
	actorstore2 "github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

type Actor = actorstore2.Actor
type CreateActorParams = actorstore2.CreateActorParams
type CreateActorRow = actorstore2.CreateActorRow
type GetActorParams = actorstore2.GetActorParams
type GetActorRow = actorstore2.GetActorRow
type GetActorByUIDParams = actorstore2.GetActorByUIDParams
type GetActorByUIDRow = actorstore2.GetActorByUIDRow
type GetActorProfileByMerchantAndEmailParams = actorstore2.GetActorProfileByMerchantAndEmailParams
type GetActorProfileByMerchantAndEmailRow = actorstore2.GetActorProfileByMerchantAndEmailRow
type ListActorsByMerchantRow = actorstore2.ListActorsByMerchantRow
type ListEmployeesByMerchantRow = actorstore2.ListEmployeesByMerchantRow
type CreateSessionParams = actorstore2.CreateSessionParams
type CreateSessionRow = actorstore2.CreateSessionRow
type GetSessionRow = actorstore2.GetSessionRow

type NewActor struct {
	MerchantID uuid.UUID
	FullName   string
	Email      string
	Password   string
}

type LoginActor struct {
	MerchantID uuid.UUID
	Email      string
	Password   string
}

type NewActorSession struct {
	RefreshTokenID        uuid.UUID
	MerchantID            uuid.UUID
	ActorID               uuid.UUID
	RefreshToken          string
	UserAgent             string
	ClientIP              string
	RefreshTokenExpiresAt time.Time
}

type Service interface {
	CreateActor(ctx context.Context, newActor NewActor) (CreateActorRow, *domainerr.DomainError)
	LoginActor(ctx context.Context, loginActor LoginActor) (GetActorRow, *domainerr.DomainError)
	CreateActorSession(ctx context.Context, newActorSession NewActorSession) (CreateSessionRow, *domainerr.DomainError)
	GetActorSession(ctx context.Context, refreshTokenID uuid.UUID) (GetSessionRow, *domainerr.DomainError)
	GetActorByUID(ctx context.Context, actorUID string) (GetActorByUIDRow, *domainerr.DomainError)
	GetActorByMerchantAndUID(ctx context.Context, merchantID uuid.UUID, actorUID string) (GetActorByUIDRow, *domainerr.DomainError)
	GetActorProfileByMerchantAndEmail(ctx context.Context, merchantID uuid.UUID, email string) (GetActorProfileByMerchantAndEmailRow, *domainerr.DomainError)
	ListActorsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]ListActorsByMerchantRow, *domainerr.DomainError)
	ListEmployeesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]ListEmployeesByMerchantRow, *domainerr.DomainError)
}
