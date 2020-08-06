package generate_test

import (
	"testing"

	"github.com/mwmahlberg/mgogenerate/generate"
)

func BenchmarkOneOf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generate.OneOf("foo", 1, []string{"a", "b", "c"})
	}
}
