package logger

import (
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMigrateLogger_Printf(t *testing.T) {
	t.Parallel()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.Logger.EXPECT().Infof("foo %v", 42)
	log := NewMigrateLogger(mock.Logger)

	// Act.
	log.Printf("foo %v", 42)
}

func TestMigrateLogger_Verbose(t *testing.T) {
	t.Parallel()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	log := NewMigrateLogger(mock.Logger)

	// Act-assert.
	require.True(t, log.Verbose())
}
