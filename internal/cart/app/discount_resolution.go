package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	pgsqlc "github.com/horiondreher/go-web-api-boilerplate/internal/merchant/store"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

func merchantDiscountFromCreateRow(row pgsqlc.CreateMerchantDiscountRow) pgsqlc.MerchantDiscount {
	return pgsqlc.MerchantDiscount{
		ID:          row.ID,
		MerchantID:  row.MerchantID,
		Type:        row.Type,
		Value:       row.Value,
		Description: row.Description,
		ValidFrom:   row.ValidFrom,
		ValidTo:     row.ValidTo,
		CreatedAt:   row.CreatedAt,
		ProductID:   row.ProductID,
		CategoryID:  row.CategoryID,
	}
}

func merchantDiscountFromGetRow(row pgsqlc.GetMerchantDiscountRow) pgsqlc.MerchantDiscount {
	return pgsqlc.MerchantDiscount{
		ID:          row.ID,
		MerchantID:  row.MerchantID,
		Type:        row.Type,
		Value:       row.Value,
		Description: row.Description,
		ValidFrom:   row.ValidFrom,
		ValidTo:     row.ValidTo,
		CreatedAt:   row.CreatedAt,
		ProductID:   row.ProductID,
		CategoryID:  row.CategoryID,
	}
}

func merchantDiscountFromListRow(row pgsqlc.ListDiscountsByMerchantRow) pgsqlc.MerchantDiscount {
	return pgsqlc.MerchantDiscount{
		ID:          row.ID,
		MerchantID:  row.MerchantID,
		Type:        row.Type,
		Value:       row.Value,
		Description: row.Description,
		ValidFrom:   row.ValidFrom,
		ValidTo:     row.ValidTo,
		CreatedAt:   row.CreatedAt,
		ProductID:   row.ProductID,
		CategoryID:  row.CategoryID,
	}
}

func isDiscountActive(discount pgsqlc.MerchantDiscount, now time.Time) bool {
	if !discount.ValidFrom.IsZero() && now.Before(discount.ValidFrom) {
		return false
	}
	if !discount.ValidTo.IsZero() && now.After(discount.ValidTo) {
		return false
	}
	return true
}

func discountPriority(discount pgsqlc.MerchantDiscount, productID uuid.UUID, categoryID uuid.UUID) (int, bool) {
	switch {
	case discount.ProductID != uuid.Nil:
		return 1, discount.ProductID == productID
	case discount.CategoryID != uuid.Nil:
		return 2, discount.CategoryID == categoryID
	default:
		return 3, true
	}
}

func (service *Service) resolveBestDiscountForProduct(ctx context.Context, merchantID uuid.UUID, productID uuid.UUID, categoryID uuid.UUID, now time.Time) (*pgsqlc.MerchantDiscount, *domainerr.DomainError) {
	discounts, err := service.store.ListDiscountsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	var best *pgsqlc.MerchantDiscount
	bestPriority := 99
	for _, row := range discounts {
		discount := merchantDiscountFromListRow(row)
		if !isDiscountActive(discount, now) {
			continue
		}

		priority, matches := discountPriority(discount, productID, categoryID)
		if !matches {
			continue
		}

		if best == nil || priority < bestPriority || (priority == bestPriority && discount.CreatedAt.After(best.CreatedAt)) {
			selected := discount
			best = &selected
			bestPriority = priority
		}
	}

	return best, nil
}
