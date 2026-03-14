package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumericFromFloat(t *testing.T) {
	n := NumericFromFloat(10)
	// pgtype.Numeric must have Int set for non-zero values
	assert.NotNil(t, n.Int, "NumericFromFloat(10) should set Int")
	back := NumericToFloat(n)
	assert.Equal(t, 10.0, back)
}
