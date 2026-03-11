package app

import (
	"context"

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
