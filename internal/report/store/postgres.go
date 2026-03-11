package store

import (
	"context"

	generated "github.com/horiondreher/go-web-api-boilerplate/internal/report/store/generated"
)

type DBTX = generated.DBTX

type Postgres struct {
	queries *generated.Queries
}

func New(db DBTX) *Postgres {
	return &Postgres{queries: generated.New(db)}
}

func (store *Postgres) GetMonthlySalesReport(ctx context.Context, arg GetMonthlySalesReportParams) (GetMonthlySalesReportRow, error) {
	return store.queries.GetMonthlySalesReport(ctx, arg)
}
