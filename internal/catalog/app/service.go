package app

import (
	"github.com/google/uuid"
	commercestore "github.com/horiondreher/go-web-api-boilerplate/internal/commerce/store"

	pkgdb "github.com/horiondreher/go-web-api-boilerplate/pkg/db"
)

type Service struct {
	db    *pkgdb.DB
	store *commercestore.Postgres
}

func NewService(db *pkgdb.DB, store *commercestore.Postgres) *Service {
	return &Service{db: db, store: store}
}

type InventoryItem struct {
	ProductID   uuid.UUID
	ProductName string
	Quantity    int32
}
