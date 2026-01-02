package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestStdLogProvider_GetStdLog(t *testing.T) {
	t.Parallel()

	// Arrange.
	log := zaptest.NewLogger(t)
	provider := NewStdLogProvider(log)

	// Act-assert.
	require.NotNil(t, provider.GetStdLog())
}
