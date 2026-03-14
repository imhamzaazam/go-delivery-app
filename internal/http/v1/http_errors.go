package v1

import (
	"net/http"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
	"github.com/horiondreher/go-web-api-boilerplate/pkg/http-tools/httputils"
)

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	httpError := domainerr.HTTPErrorBody{
		Code:   domainerr.NotFoundError,
		Errors: "The requested resource was not found",
	}

	_ = httputils.Encode(w, r, http.StatusNotFound, httpError)
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	httpError := domainerr.HTTPErrorBody{
		Code:   domainerr.MehodNotAllowedError,
		Errors: "The request method is not allowed",
	}

	_ = httputils.Encode(w, r, http.StatusMethodNotAllowed, httpError)
}
