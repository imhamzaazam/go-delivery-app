package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/internal/merchant"
)

func TestTDD_AdminAndMerchantCreationFlow(t *testing.T) {
	fx := setupCommerceFixture(t)

	newMerchant, createErr := fx.commerceService.CreateMerchantByAdmin(fx.ctx, fx.adminActorID, merchant.NewMerchant{
		Name:          "Another Merchant",
		Ntn:           "NTN-ANOTHER-001",
		Address:       "Address A",
		Category:      "restaurant",
		ContactNumber: "03444444444444",
	})
	require.Nil(t, createErr)
	require.NotEqual(t, newMerchant.ID.String(), "")

	_, forbiddenErr := fx.commerceService.CreateMerchantByAdmin(fx.ctx, fx.merchantOwnerID, merchant.NewMerchant{
		Name:          "Forbidden Merchant",
		Ntn:           "NTN-FORBIDDEN-001",
		Address:       "Address B",
		Category:      "restaurant",
		ContactNumber: "03555555555555",
	})
	require.NotNil(t, forbiddenErr)
	require.Equal(t, 403, forbiddenErr.HTTPCode)
	require.Equal(t, domainerr.UnauthorizedError, forbiddenErr.HTTPErrorBody.Code)

	employee, employeeErr := fx.commerceService.CreateEmployeeByMerchant(fx.ctx, fx.merchantOwnerID, fx.merchantID, "Shop Employee", "employee@tenant.test", "Password#123")
	require.Nil(t, employeeErr)
	require.Equal(t, "employee@tenant.test", employee.Email)
}
