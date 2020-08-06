package generate

import "math/rand"

func OneOf(t ...interface{}) interface{} {
	return t[rand.Intn(len(t))]
}
