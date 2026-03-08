package v1

import (
	"math"
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (adapter *HTTPAdapter) GetCartDetail(w http.ResponseWriter, r *http.Request, cartID string, params GetCartDetailParams) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		cart, err := adapter.readService.GetCartDetail(r.Context(), authUser.MerchantID, authUser.Email, cartID, string(params.PaymentType))
		if err != nil {
			return err
		}

		products := make([]CartProductResponse, 0, len(cart.Items))
		subtotal := 0.0
		taxableSubtotal := 0.0
		totalDiscount := 0.0
		var discountSummary *CartDiscountResponse
		for _, item := range cart.Items {
			quantity := int(item.Item.Quantity)
			productBasePrice := numericToFloat64(item.Product.BasePrice)
			productAddons := make([]CartProductAddonResponse, 0, len(item.Addons))
			addonTotal := 0.0
			for _, addon := range item.Addons {
				price := numericToFloat64(addon.Price)
				addonTotal += price * float64(quantity)
				addonID := addon.ID
				addonName := addon.Name
				addonPrice := price
				productAddons = append(productAddons, CartProductAddonResponse{
					Id:    &addonID,
					Name:  &addonName,
					Price: &addonPrice,
				})
			}

			lineDiscount := numericToFloat64(item.Item.AppliedDiscountAmount)
			totalDiscount += lineDiscount
			lineTaxableAmount := roundCartCurrency((productBasePrice * float64(quantity)) + addonTotal)
			lineSubtotal := roundCartCurrency(lineTaxableAmount - lineDiscount)
			taxableSubtotal += lineTaxableAmount
			subtotal += lineSubtotal

			if item.Discount != nil {
				if discountSummary == nil {
					discountID := item.Discount.ID
					discountType := string(item.Discount.Type)
					discountValue := numericToFloat64(item.Discount.Value)
					discountAmount := 0.0
					discountDescription := textString(item.Discount.Description)
					discountSummary = &CartDiscountResponse{
						Id:          &discountID,
						Type:        &discountType,
						Value:       &discountValue,
						Amount:      &discountAmount,
						Description: &discountDescription,
					}
				}
				discountAmount := roundCartCurrency(*discountSummary.Amount + lineDiscount)
				discountSummary.Amount = &discountAmount
			}

			productID := item.Product.ID
			productName := item.Product.Name
			productPrice := productBasePrice
			productQuantity := quantity
			products = append(products, CartProductResponse{
				Id:       &productID,
				Name:     &productName,
				Price:    &productPrice,
				Quantity: &productQuantity,
				Addons:   &productAddons,
			})
		}

		if discountSummary != nil && totalDiscount == 0 {
			discountSummary = nil
		}

		responseCartID := cart.Cart.ID
		taxRate := roundCartCurrency(cart.VatRate)
		taxAmount := roundCartCurrency(taxableSubtotal * (cart.VatRate / 100.0))
		responseTotalPrice := roundCartCurrency(subtotal + taxAmount)

		response := CartDetailResponse{
			CartId:     &responseCartID,
			TotalPrice: &responseTotalPrice,
			Discount:   discountSummary,
			Tax: &CartTaxResponse{
				Rate:   &taxRate,
				Amount: &taxAmount,
			},
			Products: &products,
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func roundCartCurrency(value float64) float64 {
	return math.Round(value*100) / 100
}
