package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	commerce "github.com/horiondreher/go-web-api-boilerplate/internal/commerce"
	commercestore "github.com/horiondreher/go-web-api-boilerplate/internal/commerce/store"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

func merchantDiscountFromGetRow(row commerce.GetMerchantDiscountRow) commerce.MerchantDiscount {
	return commercestore.ToMerchantDiscountFromGetRow(row)
}

func merchantDiscountFromListRow(row commerce.ListDiscountsByMerchantRow) commerce.MerchantDiscount {
	return commercestore.ToMerchantDiscountFromListRow(row)
}

func isDiscountActive(discount commerce.MerchantDiscount, now time.Time) bool {
	if !discount.ValidFrom.IsZero() && now.Before(discount.ValidFrom) {
		return false
	}
	if !discount.ValidTo.IsZero() && now.After(discount.ValidTo) {
		return false
	}
	return true
}

func discountPriority(discount commerce.MerchantDiscount, productID uuid.UUID, categoryID uuid.UUID) (int, bool) {
	switch {
	case discount.ProductID != uuid.Nil:
		return 1, discount.ProductID == productID
	case discount.CategoryID != uuid.Nil:
		return 2, discount.CategoryID == categoryID
	default:
		return 3, true
	}
}

func (service *Service) resolveBestDiscountForProduct(ctx context.Context, merchantID uuid.UUID, productID uuid.UUID, categoryID uuid.UUID, now time.Time) (*commerce.MerchantDiscount, *domainerr.DomainError) {
	discounts, err := service.store.ListDiscountsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, domainerr.MatchPostgresError(err)
	}

	var best *commerce.MerchantDiscount
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
