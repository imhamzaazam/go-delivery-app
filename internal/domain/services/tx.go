package services

import (
	"context"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/jackc/pgx/v5"
)

func (service *CommerceManager) runInTx(ctx context.Context, fn func(tx pgx.Tx, store *pgsqlc.Queries) *domainerr.DomainError) *domainerr.DomainError {
	tx, err := service.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainerr.NewInternalError(err)
	}

	store := pgsqlc.New(tx)
	if domainErr := fn(tx, store); domainErr != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return domainerr.NewInternalError(rollbackErr)
		}
		return domainErr
	}

	if err := tx.Commit(ctx); err != nil {
		return domainerr.NewInternalError(err)
	}

	return nil
}
