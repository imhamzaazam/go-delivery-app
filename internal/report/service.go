package report

import (
	"context"

	"github.com/google/uuid"
	reportstore "github.com/horiondreher/go-web-api-boilerplate/internal/report/store"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	coverage "github.com/horiondreher/go-web-api-boilerplate/internal/coverage"
)

type GetMonthlySalesReportParams = reportstore.GetMonthlySalesReportParams
type GetMonthlySalesReportRow = reportstore.GetMonthlySalesReportRow

type SalesReport struct {
	Month          int
	Year           int
	TotalSales     float64
	TotalTax       float64
	TotalDiscount  float64
	ProfitEstimate float64
}

type Service interface {
	GetMonthlySalesReport(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, month int, year int) (SalesReport, *domainerr.DomainError)
	ListMerchantServiceZonesByMerchant(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string) ([]coverage.ListMerchantServiceZonesByMerchantRow, *domainerr.DomainError)
}
