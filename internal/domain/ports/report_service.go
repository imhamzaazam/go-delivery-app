package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type SalesReport struct {
	Month          int
	Year           int
	TotalSales     float64
	TotalTax       float64
	TotalDiscount  float64
	ProfitEstimate float64
}

type ReportService interface {
	GetMonthlySalesReport(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, month int, year int) (SalesReport, *domainerr.DomainError)
	ListMerchantServiceZonesByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]pgsqlc.ListMerchantServiceZonesByMerchantRow, *domainerr.DomainError)
}
