package generate_test

import (
	"math"
	"testing"

	"github.com/mwmahlberg/mongen/generate"
	"github.com/stretchr/testify/assert"
)

func (s *GeneratorTestSuite) TestDecimal() {
	testCases := []struct {
		desc         string
		min          float64
		max          float64
		max_expected float64
		min_expected float64
	}{
		{
			desc:         "Min and max at borders",
			min:          math.SmallestNonzeroFloat64,
			max:          math.MaxFloat64,
			min_expected: math.SmallestNonzeroFloat64,
			max_expected: math.MaxFloat64,
		},
		{
			desc:         "Min and max 0",
			min:          0.00,
			max:          0.00,
			min_expected: math.SmallestNonzeroFloat64,
			max_expected: math.MaxFloat64,
		},
	}
	for _, tC := range testCases {
		s.T().Run(tC.desc, func(t *testing.T) {
			for i := 0; i < 50; i++ {
				v := generate.Decimal(tC.min, tC.max)
				assert.True(t, v < tC.max_expected && v > tC.min_expected, "Test failed for %f", v)
			}
		})
	}
}

func BenchmarkDecimal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generate.Decimal(float64(i), float64(i))
	}
}
