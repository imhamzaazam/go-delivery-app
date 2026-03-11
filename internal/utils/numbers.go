package utils

import (
	"math"

	"github.com/jackc/pgx/v5/pgtype"
)

func Round2(d float64) float64 {
	return math.Round(d*100) / 100
}

func NumericToFloat(n pgtype.Numeric) float64 {
	v, _ := n.Float64Value()
	return v.Float64
}

func NumericFromFloat(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	n.Scan(f)
	return n
}
