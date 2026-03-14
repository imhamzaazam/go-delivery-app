package utils

import (
	"fmt"
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
	// Scan(float64) does not populate pgtype.Numeric correctly; use string representation
	_ = n.Scan(fmt.Sprintf("%v", f))
	return n
}
