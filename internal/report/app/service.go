package app

import (
	"github.com/google/uuid"
	commercestore "github.com/horiondreher/go-web-api-boilerplate/internal/commerce/store"
	reportdomain "github.com/horiondreher/go-web-api-boilerplate/internal/report"

	pkgdb "github.com/horiondreher/go-web-api-boilerplate/pkg/db"
)

type Service struct {
	db    *pkgdb.DB
	store *commercestore.Postgres
}

func NewService(db *pkgdb.DB, store *commercestore.Postgres) *Service {
	return &Service{db: db, store: store}
}

type SalesReport = reportdomain.SalesReport

type InventoryItem struct {
	ProductID   uuid.UUID
	ProductName string
	Quantity    int32
}
