package v1

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func ptrUUID(id uuid.UUID) *openapi_types.UUID {
	value := openapi_types.UUID(id)
	return &value
}

func numericToFloat64(value pgtype.Numeric) float64 {
	number, err := value.Float64Value()
	if err != nil || !number.Valid {
		return 0
	}

	return number.Float64
}

func textString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func mustString(value interface{}) string {
	return fmt.Sprint(value)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
