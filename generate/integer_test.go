package generate_test

import (
	"testing"

	"github.com/mwmahlberg/mgogenerate/generate"
	"github.com/stretchr/testify/assert"
)

func (s *GeneratorTestSuite) TestInteger() {
	var v int64
	// Ensure that the boundaries are obeyed
	// by the way we calculate it.
	for i := 1; i < 50; i++ {
		v = generate.Integer(int64(i), int64(i))
		assert.EqualValues(s.T(), int64(i), v)
	}
}

func BenchmarkInteger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generate.Integer(int64(i), int64(i))
	}
}
