package httperr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func MatchEncodingError(err error) *domainerr.DomainError {
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return &domainerr.DomainError{
			HTTPCode:      http.StatusBadRequest,
			OriginalError: err.Error(),
			HTTPErrorBody: domainerr.HTTPErrorBody{
				Code:   domainerr.JsonDecodeError,
				Errors: "The request body is invalid",
			},
		}
	}

	jsonErr, ok := err.(*json.UnmarshalTypeError)
	if ok {
		return transformUnmarshalError(jsonErr)
	}

	if strings.Contains(err.Error(), "failed to pass regex validation") {
		return transformRegexValidationError(err)
	}

	return domainerr.NewInternalError(err)
}

func transformUnmarshalError(err *json.UnmarshalTypeError) *domainerr.DomainError {
	errors := make(map[string]string)
	errors[err.Field] = fmt.Sprintf("The field is invalid. Expected type %v", err.Type)

	return &domainerr.DomainError{
		HTTPCode:      http.StatusUnprocessableEntity,
		OriginalError: err.Error(),
		HTTPErrorBody: domainerr.HTTPErrorBody{
			Code:   domainerr.JsonDecodeError,
			Errors: errors,
		},
	}
}

func transformRegexValidationError(err error) *domainerr.DomainError {
	field := "body"
	message := "The field is invalid"

	parts := strings.SplitN(err.Error(), ":", 2)
	if len(parts) == 2 {
		field = strings.TrimSpace(parts[0])
	}

	if strings.Contains(err.Error(), "regex validation") {
		message = "The field must match the required format"
		if strings.EqualFold(field, "email") {
			message = "The field must be a valid email address"
		}
	}

	return &domainerr.DomainError{
		HTTPCode:      http.StatusUnprocessableEntity,
		OriginalError: err.Error(),
		HTTPErrorBody: domainerr.HTTPErrorBody{
			Code: domainerr.ValidationError,
			Errors: map[string]string{
				field: message,
			},
		},
	}
}
