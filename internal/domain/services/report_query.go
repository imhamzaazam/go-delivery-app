package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/pgsqlc"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/ports"
)

func (service *CommerceManager) GetMonthlySalesReport(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, month int, year int) (ports.SalesReport, *domainerr.DomainError) {
	parsedMerchantID, parseErr := parseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return ports.SalesReport{}, parseErr
	}

	if _, accessErr := service.requireMerchantViewAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return ports.SalesReport{}, accessErr
	}

	row, err := service.store.GetMonthlySalesReport(ctx, pgsqlc.GetMonthlySalesReportParams{
		MerchantID: parsedMerchantID,
		Month:      int32(month),
		Year:       int32(year),
	})
	if err != nil {
		return ports.SalesReport{}, domainerr.MatchPostgresError(err)
	}

	return ports.SalesReport{
		Month:          month,
		Year:           year,
		TotalSales:     round2(numericToFloat(row.TotalSales)),
		TotalTax:       round2(numericToFloat(row.TotalTax)),
		TotalDiscount:  round2(numericToFloat(row.TotalDiscount)),
		ProfitEstimate: round2(numericToFloat(row.ProfitEstimate)),
	}, nil
}
