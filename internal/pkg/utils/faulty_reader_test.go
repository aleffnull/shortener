package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFaultyReader(t *testing.T) {
	t.Parallel()

	// Act-assert.
	count, err := AFaultyReader.Read([]byte("foo"))
	require.Zero(t, count)
	require.Error(t, err)
}
