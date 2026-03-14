package utils

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func PtrUUID(id uuid.UUID) *openapi_types.UUID {
	value := openapi_types.UUID(id)
	return &value
}

func NumericToFloat64(value pgtype.Numeric) float64 {
	number, err := value.Float64Value()
	if err != nil || !number.Valid {
		return 0
	}

	return number.Float64
}

func TextString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}