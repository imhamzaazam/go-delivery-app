package services

import (
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgtype"
)

func numericFromFloat(value float64) pgtype.Numeric {
	rounded := round2(value)
	numeric := pgtype.Numeric{}
	_ = numeric.Scan(fmt.Sprintf("%.2f", rounded))
	return numeric
}

func numericToFloat(value pgtype.Numeric) float64 {
	number, err := value.Float64Value()
	if err != nil || !number.Valid {
		return 0
	}
	return number.Float64
}

func textValue(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
