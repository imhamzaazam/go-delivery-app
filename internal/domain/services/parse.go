package services

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
)

func parseUUID(value string, field string) (uuid.UUID, *domainerr.DomainError) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, domainerr.NewDomainError(http.StatusBadRequest, domainerr.ValidationError, fmt.Sprintf("invalid %s", field), err)
	}

	return parsed, nil
}
