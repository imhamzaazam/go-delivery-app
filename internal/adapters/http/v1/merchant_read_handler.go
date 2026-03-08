package v1

import (
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (adapter *HTTPAdapter) ListBranchesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		branches, err := adapter.merchantService.ListBranchesByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]BranchResponse, 0, len(branches))
		for _, branch := range branches {
			contactNumber := branch.ContactNumber.String
			city := string(branch.City)
			createdAt := branch.CreatedAt
			updatedAt := branch.UpdatedAt
			item := BranchResponse{
				Id:            ptrUUID(branch.ID),
				MerchantId:    ptrUUID(branch.MerchantID),
				Name:          &branch.Name,
				Address:       &branch.Address,
				ContactNumber: &contactNumber,
				City:          &city,
				CreatedAt:     &createdAt,
				UpdatedAt:     &updatedAt,
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) ListDiscountsByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		discounts, err := adapter.merchantService.ListDiscountsByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]DiscountResponse, 0, len(discounts))
		for _, discount := range discounts {
			description := discount.Description.String
			discountType := string(discount.Type)
			value := numericToFloat64(discount.Value)
			createdAt := discount.CreatedAt
			item := DiscountResponse{
				Id:          ptrUUID(discount.ID),
				MerchantId:  ptrUUID(discount.MerchantID),
				Type:        &discountType,
				Value:       &value,
				Description: &description,
				CreatedAt:   &createdAt,
			}
			if !discount.ValidFrom.IsZero() {
				validFrom := discount.ValidFrom
				item.ValidFrom = &validFrom
			}
			if !discount.ValidTo.IsZero() {
				validTo := discount.ValidTo
				item.ValidTo = &validTo
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}

func (adapter *HTTPAdapter) ListRolesByMerchant(w http.ResponseWriter, r *http.Request) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		roles, err := adapter.merchantService.ListRolesByMerchant(r.Context(), authUser.MerchantID.String())
		if err != nil {
			return err
		}

		response := make([]RoleResponse, 0, len(roles))
		for _, role := range roles {
			description := role.Description.String
			roleType := string(role.RoleType)
			createdAt := role.CreatedAt
			item := RoleResponse{
				Id:          ptrUUID(role.ID),
				MerchantId:  ptrUUID(role.MerchantID),
				RoleType:    &roleType,
				Description: &description,
				CreatedAt:   &createdAt,
			}
			response = append(response, item)
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}
