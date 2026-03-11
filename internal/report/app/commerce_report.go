package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/utils"
)

func (service *Service) SalesReportByMonth(ctx context.Context, viewerActorID uuid.UUID, merchantID uuid.UUID, month int, year int) (SalesReport, *domainerr.DomainError) {
	allowed, allowErr := service.canViewMerchant(ctx, viewerActorID, merchantID)
	if allowErr != nil {
		return SalesReport{}, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return SalesReport{}, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "admin or merchant role required", fmt.Errorf("admin or merchant role required"))
	}

	report := SalesReport{Month: month, Year: year}
	err := service.db.QueryRow(ctx, `
        SELECT
            COALESCE(SUM(oi.line_total), 0) AS total_sales,
            COALESCE(SUM(oi.tax_amount), 0) AS total_tax,
            COALESCE(SUM(oi.discount_amount), 0) AS total_discount,
            COALESCE(SUM(oi.base_amount + oi.addon_amount - oi.discount_amount), 0) AS profit_estimate
        FROM orders o
        JOIN order_items oi ON oi.order_id = o.id
        WHERE o.merchant_id = $1
          AND o.status IN ('accepted', 'out_for_delivery', 'delivered')
          AND EXTRACT(MONTH FROM o.created_at) = $2
          AND EXTRACT(YEAR FROM o.created_at) = $3
    `, merchantID, month, year).Scan(&report.TotalSales, &report.TotalTax, &report.TotalDiscount, &report.ProfitEstimate)
	if err != nil {
		return SalesReport{}, domainerr.NewInternalError(err)
	}

	report.TotalSales = utils.Round2(report.TotalSales)
	report.TotalTax = utils.Round2(report.TotalTax)
	report.TotalDiscount = utils.Round2(report.TotalDiscount)
	report.ProfitEstimate = utils.Round2(report.ProfitEstimate)

	return report, nil
}

func (service *Service) ViewInventory(ctx context.Context, viewerActorID uuid.UUID, merchantID uuid.UUID) ([]InventoryItem, *domainerr.DomainError) {
	allowed, allowErr := service.canViewMerchant(ctx, viewerActorID, merchantID)
	if allowErr != nil {
		return nil, domainerr.NewInternalError(allowErr)
	}
	if !allowed {
		return nil, domainerr.NewDomainError(http.StatusForbidden, domainerr.UnauthorizedError, "admin or merchant role required", fmt.Errorf("admin or merchant role required"))
	}

	rows, err := service.db.Query(ctx, `
        SELECT p.id, p.name, COALESCE(SUM(pi.quantity), 0) AS quantity
        FROM products p
        LEFT JOIN product_inventory pi ON pi.product_id = p.id
        WHERE p.merchant_id = $1
        GROUP BY p.id, p.name
        ORDER BY p.name
    `, merchantID)
	if err != nil {
		return nil, domainerr.NewInternalError(err)
	}
	defer rows.Close()

	result := make([]InventoryItem, 0)
	for rows.Next() {
		var item InventoryItem
		if scanErr := rows.Scan(&item.ProductID, &item.ProductName, &item.Quantity); scanErr != nil {
			return nil, domainerr.NewInternalError(scanErr)
		}
		result = append(result, item)
	}

	if rows.Err() != nil {
		return nil, domainerr.NewInternalError(rows.Err())
	}

	return result, nil
}
