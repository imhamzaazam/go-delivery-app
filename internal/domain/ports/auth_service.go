package ports

import (
	"time"

	"github.com/google/uuid"
)

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
