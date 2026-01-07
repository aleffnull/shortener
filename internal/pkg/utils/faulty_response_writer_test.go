package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFaultyResponseWriter_WriteHeader(t *testing.T) {
	t.Parallel()

	// Arrange.
	writer := NewFaultyResponseWriter()

	// Act.
	writer.WriteHeader(http.StatusInternalServerError)

	// Assert.
	require.Equal(t, http.StatusInternalServerError, writer.StatusCode)
	require.Equal(t, http.Header{}, writer.Header())
}

func TestFaultyResponseWriter_Write(t *testing.T) {
	t.Parallel()

	// Arrange.
	writer := NewFaultyResponseWriter()

	// Act.
	count, err := writer.Write([]byte("foo"))

	// Assert.
	require.Zero(t, count)
	require.Error(t, err)
}
