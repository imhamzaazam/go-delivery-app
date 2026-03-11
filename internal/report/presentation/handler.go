package presentation

import (
	"net/http"

	api "github.com/horiondreher/go-web-api-boilerplate/api"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core"
	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

type Handler struct {
	shared *core.Shared
}

func New(shared *core.Shared) *Handler {
	return &Handler{shared: shared}
}

func (handler *Handler) GetMonthlySalesReport(w http.ResponseWriter, r *http.Request, params api.GetMonthlySalesReportParams) {
	handler.shared.ServeAuthenticated(w, r, func(w http.ResponseWriter, r *http.Request) *domainerr.DomainError {
		authUser, authErr := handler.shared.CurrentAuthUser(r)
		if authErr != nil {
			return authErr
		}

		report, err := handler.shared.ReadService.GetMonthlySalesReport(r.Context(), authUser.MerchantID, authUser.Email, authUser.MerchantID.String(), params.Month, params.Year)
		if err != nil {
			return err
		}

		response := api.SalesReportResponse{
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
