package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	actorstore2 "github.com/horiondreher/go-web-api-boilerplate/internal/actor/store/generated"
	"github.com/jackc/pgx/v5/pgtype"

	actor "github.com/horiondreher/go-web-api-boilerplate/internal/actor"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

type actorStore interface {
	CreateActor(ctx context.Context, arg actorstore2.CreateActorParams) (actorstore2.CreateActorRow, error)
	GetActor(ctx context.Context, arg actorstore2.GetActorParams) (actorstore2.GetActorRow, error)
	CreateSession(ctx context.Context, arg actorstore2.CreateSessionParams) (actorstore2.CreateSessionRow, error)
	GetSession(ctx context.Context, id uuid.UUID) (actorstore2.GetSessionRow, error)
	GetActorByUID(ctx context.Context, arg actorstore2.GetActorByUIDParams) (actorstore2.GetActorByUIDRow, error)
	GetActorProfileByMerchantAndEmail(ctx context.Context, arg actorstore2.GetActorProfileByMerchantAndEmailParams) (actorstore2.GetActorProfileByMerchantAndEmailRow, error)
	ListActorsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]actorstore2.ListActorsByMerchantRow, error)
	ListEmployeesByMerchant(ctx context.Context, merchantID uuid.UUID) ([]actorstore2.ListEmployeesByMerchantRow, error)
}

type Service struct {
	store actorStore
}

func NewService(store actorStore) *Service {
	return &Service{
		store: store,
	}
}

func (service *Service) CreateActor(ctx context.Context, newActor actor.NewActor) (actorstore2.CreateActorRow, *domainerr.DomainError) {
	hashedPassword, hashErr := utils.HashPasswordOrNoop(newActor.Password)
	if hashErr != nil {
		return actorstore2.CreateActorRow{}, hashErr
	}

	fullNameParts := strings.Fields(strings.TrimSpace(newActor.FullName))
	firstName := ""
	lastName := ""
	if len(fullNameParts) > 0 {
		firstName = fullNameParts[0]
	}
	if len(fullNameParts) > 1 {
		lastName = strings.Join(fullNameParts[1:], " ")
	}

	args := actorstore2.CreateActorParams{
		MerchantID:   newActor.MerchantID,
		Email:        newActor.Email,
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		IsActive:     true,
		LastLogin:    time.Now(),
	}

	user, err := service.store.CreateActor(ctx, args)
	if err != nil {
		return actorstore2.CreateActorRow{}, domainerr.MatchPostgresError(err)
	}

	return user, nil
}

func (service *Service) LoginActor(ctx context.Context, loginActor actor.LoginActor) (actorstore2.GetActorRow, *domainerr.DomainError) {
	user, err := service.store.GetActor(ctx, actorstore2.GetActorParams{
		MerchantID: loginActor.MerchantID,
		Email:      loginActor.Email,
	})
	if err != nil {
		return actorstore2.GetActorRow{}, domainerr.MatchPostgresError(err)
	}
	if !user.IsActive {
		return actorstore2.GetActorRow{}, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "actor is inactive", fmt.Errorf("actor is inactive"))
	}

	passErr := utils.CheckPassword(loginActor.Password, user.Password)
	if passErr != nil {
		return actorstore2.GetActorRow{}, passErr
	}

	return user, nil
}

func (service *Service) CreateActorSession(ctx context.Context, newActorSession actor.NewActorSession) (actorstore2.CreateSessionRow, *domainerr.DomainError) {
	session, err := service.store.CreateSession(ctx, actorstore2.CreateSessionParams{
		ID:           newActorSession.RefreshTokenID,
		MerchantID:   newActorSession.MerchantID,
		ActorID:      pgtype.UUID{Bytes: newActorSession.ActorID, Valid: true},
		RefreshToken: newActorSession.RefreshToken,
		ExpiresAt:    newActorSession.RefreshTokenExpiresAt,
		UserAgent:    newActorSession.UserAgent,
		ClientIP:     newActorSession.ClientIP,
	})
	if err != nil {
		return actorstore2.CreateSessionRow{}, domainerr.MatchPostgresError(err)
	}

	return session, nil
}

func (service *Service) GetActorSession(ctx context.Context, refreshTokenID uuid.UUID) (actorstore2.GetSessionRow, *domainerr.DomainError) {
	session, err := service.store.GetSession(ctx, refreshTokenID)
	if err != nil {
		return actorstore2.GetSessionRow{}, domainerr.MatchPostgresError(err)
	}

	return session, nil
}

func (service *Service) GetActorByUID(ctx context.Context, actorUID string) (actorstore2.GetActorByUIDRow, *domainerr.DomainError) {
	return service.GetActorByMerchantAndUID(ctx, uuid.Nil, actorUID)
}

func (service *Service) GetActorByMerchantAndUID(ctx context.Context, merchantID uuid.UUID, actorUID string) (actorstore2.GetActorByUIDRow, *domainerr.DomainError) {
	parsedUID, err := uuid.Parse(actorUID)
	if err != nil {
		return actorstore2.GetActorByUIDRow{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid actor uid", err)
	}

	user, err := service.store.GetActorByUID(ctx, actorstore2.GetActorByUIDParams{
		MerchantID: merchantID,
		ID:         parsedUID,
	})
	if err != nil {
		return actorstore2.GetActorByUIDRow{}, domainerr.MatchPostgresError(err)
	}

	return user, nil
}
