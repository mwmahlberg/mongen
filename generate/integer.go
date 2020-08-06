package generate

import "math/rand"

// Integer generates a random int64 between min and max.
func Integer(min, max int64) int64 {
	return rand.Int63n(max-min+1) + min
}
