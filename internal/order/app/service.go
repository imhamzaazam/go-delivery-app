package app

import (
	"context"

	"github.com/google/uuid"
	commercestore "github.com/horiondreher/go-web-api-boilerplate/internal/commerce/store"
	"github.com/jackc/pgx/v5"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	pkgdb "github.com/horiondreher/go-web-api-boilerplate/pkg/db"
)

type Service struct {
	db    *pkgdb.DB
	store *commercestore.Postgres
}

func NewService(db *pkgdb.DB, store *commercestore.Postgres) *Service {
	return &Service{db: db, store: store}
}

type OrderLineBill struct {
	ProductID      uuid.UUID
	ProductName    string
	BasePrice      float64
	PaymentMethod  string
	Quantity       int32
	BaseAmount     float64
	AddonAmount    float64
	DiscountAmount float64
	FinalPrice     float64
	TaxAmount      float64
	Vat            float64
	LineTotal      float64
	TotalPrice     float64
}

type OrderBill struct {
	OrderID     uuid.UUID
	PaymentType string
	VatRate     float64
	Subtotal    float64
	TotalTax    float64
	Total       float64
	LineItems   []OrderLineBill
}

func (service *Service) runInTx(ctx context.Context, fn func(tx pgx.Tx, store *commercestore.Postgres) *domainerr.DomainError) *domainerr.DomainError {
	tx, err := service.db.Pool.Begin(ctx)
	if err != nil {
		return domainerr.NewInternalError(err)
	}
	defer tx.Rollback(ctx)
	store := commercestore.New(tx)
	if err := fn(tx, store); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return domainerr.NewInternalError(err)
	}
	return nil
}
