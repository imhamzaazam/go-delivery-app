package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type ActorManager struct {
	store pgsqlc.Querier
}

func NewActorManager(store pgsqlc.Querier) *ActorManager {
	return &ActorManager{
		store: store,
	}
}

func (service *ActorManager) CreateActor(ctx context.Context, newActor ports.NewActor) (pgsqlc.CreateActorRow, *domainerr.DomainError) {
	hashedPassword, hashErr := utils.HashPasswordOrNoop(newActor.Password)
	if hashErr != nil {
		return pgsqlc.CreateActorRow{}, hashErr
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

	args := pgsqlc.CreateActorParams{
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
		return pgsqlc.CreateActorRow{}, domainerr.MatchPostgresError(err)
	}

	return user, nil
}

func (service *ActorManager) LoginActor(ctx context.Context, loginActor ports.LoginActor) (pgsqlc.GetActorRow, *domainerr.DomainError) {
	user, err := service.store.GetActor(ctx, pgsqlc.GetActorParams{
		MerchantID: loginActor.MerchantID,
		Email:      loginActor.Email,
	})
	if err != nil {
		return pgsqlc.GetActorRow{}, domainerr.MatchPostgresError(err)
	}
	if !user.IsActive {
		return pgsqlc.GetActorRow{}, domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "actor is inactive", fmt.Errorf("actor is inactive"))
	}

	passErr := utils.CheckPassword(loginActor.Password, user.Password)
	if passErr != nil {
		return pgsqlc.GetActorRow{}, passErr
	}

	return user, nil
}

func (service *ActorManager) CreateActorSession(ctx context.Context, newActorSession ports.NewActorSession) (pgsqlc.CreateSessionRow, *domainerr.DomainError) {
	session, err := service.store.CreateSession(ctx, pgsqlc.CreateSessionParams{
		ID:           newActorSession.RefreshTokenID,
		MerchantID:   newActorSession.MerchantID,
		ActorID:      pgtype.UUID{Bytes: newActorSession.ActorID, Valid: true},
		RefreshToken: newActorSession.RefreshToken,
		ExpiresAt:    newActorSession.RefreshTokenExpiresAt,
		UserAgent:    newActorSession.UserAgent,
		ClientIP:     newActorSession.ClientIP,
	})
	if err != nil {
		return pgsqlc.CreateSessionRow{}, domainerr.MatchPostgresError(err)
	}

	return session, nil
}

func (service *ActorManager) GetActorSession(ctx context.Context, refreshTokenID uuid.UUID) (pgsqlc.GetSessionRow, *domainerr.DomainError) {
	session, err := service.store.GetSession(ctx, refreshTokenID)
	if err != nil {
		return pgsqlc.GetSessionRow{}, domainerr.MatchPostgresError(err)
	}

	return session, nil
}

func (service *ActorManager) GetActorByUID(ctx context.Context, actorUID string) (pgsqlc.GetActorByUIDRow, *domainerr.DomainError) {
	return service.GetActorByMerchantAndUID(ctx, uuid.Nil, actorUID)
}

func (service *ActorManager) GetActorByMerchantAndUID(ctx context.Context, merchantID uuid.UUID, actorUID string) (pgsqlc.GetActorByUIDRow, *domainerr.DomainError) {
	parsedUID, err := uuid.Parse(actorUID)
	if err != nil {
		return pgsqlc.GetActorByUIDRow{}, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid actor uid", err)
	}

	user, err := service.store.GetActorByUID(ctx, pgsqlc.GetActorByUIDParams{
		MerchantID: merchantID,
		ID:         parsedUID,
	})
	if err != nil {
		return pgsqlc.GetActorByUIDRow{}, domainerr.MatchPostgresError(err)
	}

	return user, nil
}
