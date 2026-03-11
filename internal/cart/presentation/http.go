package presentation

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/google/uuid"
	cartstore "github.com/horiondreher/go-web-api-boilerplate/internal/cart/store"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

type Handler struct {
	shared *core.Shared
}

type createCartValidation struct {
	BranchID string `validate:"required,uuid"`
	CartID   string `validate:"required,uuid"`
}

type addCartItemValidation struct {
	ProductID string   `validate:"required,uuid"`
	Quantity  int      `validate:"required,gt=0"`
	AddonIDs  []string `validate:"omitempty,dive,uuid"`
}

type updateCartItemValidation struct {
	ItemID   string `validate:"required,uuid"`
	Quantity int    `validate:"required,gt=0"`
}

type removeCartItemValidation struct {
	ItemID string `validate:"required,uuid"`
}

func New(shared *core.Shared) *Handler {
	return &Handler{shared: shared}
}

func cartPreviewVATRate(paymentMethod string) (float64, bool) {
	switch strings.ToLower(strings.TrimSpace(paymentMethod)) {
	case "card":
		return 8, true
	case "cash":
		return 15, true
	default:
		return 0, false
	}
}

func cartCreatedResponse(cart cartstore.Cart) api.CreateCartResponse {
	createdAt := cart.CreatedAt
	updatedAt := cart.UpdatedAt
	response := api.CreateCartResponse{
		Id:        core.PtrUUID(cart.ID),
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}
	if cart.BranchID != uuid.Nil {
		branchID := openapi_types.UUID(cart.BranchID)
		response.BranchId = &branchID
	}
	return response
}

func cartItemCreatedResponse(item cartstore.CartItem) api.CartItemResponse {
	itemID := item.ID
	quantity := int(item.Quantity)
	return api.CartItemResponse{
		ItemId:    core.PtrUUID(itemID),
		ProductId: core.PtrUUID(item.ProductID),
		Quantity:  &quantity,
	}
}

func (handler *Handler) GetCartDetail(w http.ResponseWriter, r *http.Request, cartID string, params api.GetCartDetailParams) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		cart, err := handler.shared.ReadService.GetCartDetail(r.Context(), cartID)
		if err != nil {
			return err
		}

		vatRate := 0.0
		paymentMethod := ""
		previewEnabled := false
		if params.PaymentMethod != nil {
			paymentMethod = string(*params.PaymentMethod)
			resolvedRate, ok := cartPreviewVATRate(paymentMethod)
			if !ok {
				return domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, "invalid payment method", fmt.Errorf("invalid payment method"))
			}
			vatRate = resolvedRate
			previewEnabled = true
		}

		products := make([]api.CartProductResponse, 0, len(cart.Items))
		subtotal := 0.0
		totalVAT := 0.0
		totalDiscount := 0.0
		var discountSummary *api.CartDiscountResponse
		for _, item := range cart.Items {
			quantity := int(item.Item.Quantity)
			productBasePrice := core.NumericToFloat64(item.Product.BasePrice)
			productAddons := make([]api.CartProductAddonResponse, 0, len(item.Addons))
			addonTotal := 0.0
			for _, addon := range item.Addons {
				price := core.NumericToFloat64(addon.Price)
				addonTotal += price * float64(quantity)
				addonID := addon.ID
				addonName := addon.Name
				addonPrice := price
				productAddons = append(productAddons, api.CartProductAddonResponse{
					Id:    &addonID,
					Name:  &addonName,
					Price: &addonPrice,
				})
			}

			lineDiscount := core.NumericToFloat64(item.Item.AppliedDiscountAmount)
			totalDiscount += lineDiscount
			lineSubtotal := roundCartCurrency((productBasePrice * float64(quantity)) + addonTotal - lineDiscount)
			lineVAT := 0.0
			lineTotalPrice := lineSubtotal
			if previewEnabled {
				lineVAT = roundCartCurrency(lineSubtotal * (vatRate / 100.0))
				lineTotalPrice = roundCartCurrency(lineSubtotal + lineVAT)
				totalVAT += lineVAT
			}
			subtotal += lineSubtotal

			if item.Discount != nil {
				if discountSummary == nil {
					discountID := item.Discount.ID
					discountType := string(item.Discount.Type)
					discountValue := core.NumericToFloat64(item.Discount.Value)
					discountAmount := 0.0
					discountDescription := core.TextString(item.Discount.Description)
					discountSummary = &api.CartDiscountResponse{
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
			itemID := item.Item.ID
			productName := item.Product.Name
			productPrice := productBasePrice
			productQuantity := quantity
			productResponse := api.CartProductResponse{
				Id:       &productID,
				ItemId:   &itemID,
				Name:     &productName,
				Price:    &productPrice,
				Quantity: &productQuantity,
				Addons:   &productAddons,
			}
			if previewEnabled {
				finalPrice := lineSubtotal
				vat := lineVAT
				totalPrice := lineTotalPrice
				productResponse.FinalPrice = &finalPrice
				productResponse.Vat = &vat
				productResponse.TotalPrice = &totalPrice
			}
			products = append(products, productResponse)
		}

		if discountSummary != nil && totalDiscount == 0 {
			discountSummary = nil
		}

		responseCartID := cart.Cart.ID
		responseTotalPrice := roundCartCurrency(subtotal + totalVAT)

		response := api.CartDetailResponse{
			CartId:     &responseCartID,
			TotalPrice: &responseTotalPrice,
			Discount:   discountSummary,
			Products:   &products,
		}
		if previewEnabled {
			previewPaymentMethod := paymentMethod
			previewVatRate := vatRate
			previewSubtotal := roundCartCurrency(subtotal)
			previewTotalVAT := roundCartCurrency(totalVAT)
			response.PaymentMethod = &previewPaymentMethod
			response.VatRate = &previewVatRate
			response.Subtotal = &previewSubtotal
			response.TotalVat = &previewTotalVAT
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})(w, r)
}

func (handler *Handler) CreateCart(w http.ResponseWriter, r *http.Request) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.CreateCartRequest](r)
		if err != nil {
			return err
		}
		merchantID, merchantErr := handler.shared.CurrentMerchantID(r)
		if merchantErr != nil {
			return merchantErr
		}
		validationErr := handler.shared.Validate.Struct(createCartValidation{
			BranchID: requestBody.BranchId.String(),
			CartID:   requestBody.CartId.String(),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		cart, createErr := handler.shared.CommerceService.CreateCartHTTP(r.Context(), merchantID.String(), requestBody.BranchId.String(), "", requestBody.CartId.String())
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, cartCreatedResponse(cart))
	})(w, r)
}

func (handler *Handler) AddItemToCart(w http.ResponseWriter, r *http.Request, cartID string) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.AddCartItemRequest](r)
		if err != nil {
			return err
		}

		addonIDs := make([]string, 0)
		if requestBody.AddonIds != nil {
			addonIDs = make([]string, 0, len(*requestBody.AddonIds))
			for _, addonID := range *requestBody.AddonIds {
				addonIDs = append(addonIDs, addonID.String())
			}
		}
		validationErr := handler.shared.Validate.Struct(addCartItemValidation{
			ProductID: requestBody.ProductId.String(),
			Quantity:  requestBody.Quantity,
			AddonIDs:  addonIDs,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		discountID := ""
		if requestBody.DiscountId != nil {
			discountID = requestBody.DiscountId.String()
		}
		item, addErr := handler.shared.CommerceService.AddItemToCartHTTP(r.Context(), cartID, requestBody.ProductId.String(), int32(requestBody.Quantity), addonIDs, discountID)
		if addErr != nil {
			return addErr
		}

		return httputils.Encode(w, r, http.StatusCreated, cartItemCreatedResponse(item))
	})(w, r)
}

func (handler *Handler) UpdateCartItem(w http.ResponseWriter, r *http.Request, cartID string, itemID string) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[api.UpdateCartItemRequest](r)
		if err != nil {
			return err
		}

		validationErr := handler.shared.Validate.Struct(updateCartItemValidation{
			ItemID:   itemID,
			Quantity: requestBody.Quantity,
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		item, updateErr := handler.shared.CommerceService.UpdateCartItemQuantityHTTP(r.Context(), cartID, itemID, int32(requestBody.Quantity))
		if updateErr != nil {
			return updateErr
		}

		return httputils.Encode(w, r, http.StatusOK, cartItemCreatedResponse(item))
	})(w, r)
}

func (handler *Handler) DeleteCartItem(w http.ResponseWriter, r *http.Request, cartID string, itemID string) {
	handler.shared.Wrap(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		validationErr := handler.shared.Validate.Struct(removeCartItemValidation{ItemID: itemID})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		deleteErr := handler.shared.CommerceService.RemoveItemFromCartHTTP(r.Context(), cartID, itemID)
		if deleteErr != nil {
			return deleteErr
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})(w, r)
}

func roundCartCurrency(value float64) float64 {
	return math.Round(value*100) / 100
}
