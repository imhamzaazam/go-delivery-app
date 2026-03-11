package domainerr

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func MatchPostgresError(err error) *DomainError {
	if errors.Is(err, pgx.ErrNoRows) {
		return &DomainError{
			HTTPCode:      http.StatusNotFound,
			OriginalError: err.Error(),
			HTTPErrorBody: HTTPErrorBody{
				Code:   NotFoundError,
				Errors: "Not found",
			},
		}
	}

	pgErr, ok := err.(*pgconn.PgError)
	if ok {
		return TransformPostgresError(pgErr)
	}

	return NewInternalError(err)
}

func TransformPostgresError(err *pgconn.PgError) *DomainError {
	httpError := &DomainError{
		HTTPCode:      http.StatusBadRequest,
		OriginalError: err.Error(),
		HTTPErrorBody: HTTPErrorBody{
			Code:   QueryError,
			Errors: err.ConstraintName,
		},
	}

	switch err.Code {
	case "23505":
		httpError.HTTPCode = http.StatusConflict
		httpError.HTTPErrorBody.Code = DuplicateError
		httpError.HTTPErrorBody.Errors = MapDuplicateError(err.ConstraintName)
	}

	return httpError
}

func MapDuplicateError(constraintName string) string {
	var errorMessage string

	switch constraintName {
	case "user_email_idx":
		fallthrough
	case "actors_email_key":
		fallthrough
	case "actors_merchant_id_email_key":
		errorMessage = "The email is already in use"
	}

	return errorMessage
}
