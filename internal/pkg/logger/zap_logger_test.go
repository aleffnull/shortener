package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestZapLogger_Infof(t *testing.T) {
	t.Parallel()

	// Arrange.
	zapLog := zaptest.NewLogger(t)
	log := NewZapLogger(zapLog)

	// Act.
	log.Infof("foo %v", 42)
}

func TestZapLogger_Warnf(t *testing.T) {
	t.Parallel()

	// Arrange.
	zapLog := zaptest.NewLogger(t)
	log := NewZapLogger(zapLog)

	// Act.
	log.Warnf("foo %v", 42)
}

func TestZapLogger_Errorf(t *testing.T) {
	t.Parallel()

	// Arrange.
	zapLog := zaptest.NewLogger(t)
	log := NewZapLogger(zapLog)

	// Act.
	log.Errorf("foo %v", 42)
}

func TestZapLogger_Fatalf(t *testing.T) {
	t.Parallel()

	// Arrange.
	zapLog := zaptest.NewLogger(t, zaptest.WrapOptions(zap.WithFatalHook(zapcore.WriteThenPanic)))
	log := NewZapLogger(zapLog)

	// Act-assert.
	require.Panics(t, func() {
		log.Fatalf("foo %v", 42)
	})
}
