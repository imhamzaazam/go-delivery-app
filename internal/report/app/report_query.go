package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pgsqlc "github.com/horiondreher/go-web-api-boilerplate/internal/report/store"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) GetMonthlySalesReport(ctx context.Context, viewerMerchantID uuid.UUID, viewerEmail string, merchantID string, month int, year int) (SalesReport, *domainerr.DomainError) {
	parsedMerchantID, parseErr := utils.ParseUUID(merchantID, "merchant id")
	if parseErr != nil {
		return SalesReport{}, parseErr
	}

	if _, accessErr := service.requireMerchantViewAccess(ctx, viewerMerchantID, viewerEmail, parsedMerchantID); accessErr != nil {
		return SalesReport{}, accessErr
	}

	if periodErr := validateReportPeriod(month, year); periodErr != nil {
		return SalesReport{}, periodErr
	}

	row, err := service.store.GetMonthlySalesReport(ctx, pgsqlc.GetMonthlySalesReportParams{
		MerchantID: parsedMerchantID,
		Month:      int32(month),
		Year:       int32(year),
	})
	if err != nil {
		return SalesReport{}, domainerr.MatchPostgresError(err)
	}

	return SalesReport{
		Month:          month,
		Year:           year,
		TotalSales:     utils.Round2(utils.NumericToFloat(row.TotalSales)),
		TotalTax:       utils.Round2(utils.NumericToFloat(row.TotalTax)),
		TotalDiscount:  utils.Round2(utils.NumericToFloat(row.TotalDiscount)),
		ProfitEstimate: utils.Round2(utils.NumericToFloat(row.ProfitEstimate)),
	}, nil
}

func validateReportPeriod(month int, year int) *domainerr.DomainError {
	if month < 1 || month > 12 {
		return domainerr.NewDomainError(400, domainerr.ValidationError, "invalid month", fmt.Errorf("month must be between 1 and 12"))
	}
	if year < 2000 || year > 9999 {
		return domainerr.NewDomainError(400, domainerr.ValidationError, "invalid year", fmt.Errorf("year must be between 2000 and 9999"))
	}
	return nil
}
