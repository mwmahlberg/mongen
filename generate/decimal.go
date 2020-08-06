package generate

import "math/rand"

func Decimal(min, max float64) float64 {

	if min == 0 && max == 0 {
		return rand.Float64()
	}

	return rand.Float64()*(max-min) + min
}
