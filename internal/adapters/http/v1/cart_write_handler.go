package v1

import (
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httperr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

type createCartValidation struct {
	MerchantID string `validate:"required,uuid"`
	BranchID   string `validate:"required,uuid"`
	CartID     string `validate:"required,uuid"`
}

type addCartItemValidation struct {
	ProductID string   `validate:"required,uuid"`
	Quantity  int      `validate:"required,gt=0"`
	AddonIDs  []string `validate:"omitempty,dive,uuid"`
}

func (adapter *HTTPAdapter) CreateCart(w http.ResponseWriter, r *http.Request) {
	adapter.handlerWrapper(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[CreateCartRequest](r)
		if err != nil {
			return err
		}
		validationErr := validate.Struct(createCartValidation{
			MerchantID: requestBody.MerchantId.String(),
			BranchID:   requestBody.BranchId.String(),
			CartID:     requestBody.CartId.String(),
		})
		if validationErr != nil {
			return httperr.MatchValidationError(validationErr)
		}

		cart, createErr := adapter.commerceService.CreateCartHTTP(r.Context(), requestBody.MerchantId.String(), requestBody.BranchId.String(), "", requestBody.CartId.String())
		if createErr != nil {
			return createErr
		}

		return httputils.Encode(w, r, http.StatusCreated, cartCreatedResponse(cart))
	})(w, r)
}

func (adapter *HTTPAdapter) AddItemToCart(w http.ResponseWriter, r *http.Request, cartID string) {
	adapter.handlerWrapper(func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		requestBody, err := httputils.Decode[AddCartItemRequest](r)
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
		validationErr := validate.Struct(addCartItemValidation{
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
		item, addErr := adapter.commerceService.AddItemToCartHTTP(r.Context(), cartID, requestBody.ProductId.String(), int32(requestBody.Quantity), addonIDs, discountID)
		if addErr != nil {
			return addErr
		}

		return httputils.Encode(w, r, http.StatusCreated, cartItemCreatedResponse(item))
	})(w, r)
}
