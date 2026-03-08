package v1

import (
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/adapters/http/httputils"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func (adapter *HTTPAdapter) GetMonthlySalesReport(w http.ResponseWriter, r *http.Request, params GetMonthlySalesReportParams) {
	adapter.serveAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := adapter.currentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		report, err := adapter.readService.GetMonthlySalesReport(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), params.Month, params.Year)
		if err != nil {
			return err
		}

		response := SalesReportResponse{
			Month:          &report.Month,
			Year:           &report.Year,
			TotalSales:     &report.TotalSales,
			TotalTax:       &report.TotalTax,
			TotalDiscount:  &report.TotalDiscount,
			ProfitEstimate: &report.ProfitEstimate,
		}

		return httputils.Encode(w, r, http.StatusOK, response)
	})
}
