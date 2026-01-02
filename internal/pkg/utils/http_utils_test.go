package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandleServerError(t *testing.T) {
	t.Parallel()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.Logger.EXPECT().Errorf(gomock.Any(), assert.AnError)
	response := httptest.NewRecorder()

	// Act.
	HandleServerError(response, assert.AnError, mock.Logger)

	// Assert.
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestHandleRequestError(t *testing.T) {
	t.Parallel()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.Logger.EXPECT().Errorf(gomock.Any(), assert.AnError)
	response := httptest.NewRecorder()

	// Act.
	HandleRequestError(response, assert.AnError, mock.Logger)

	// Assert.
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestHandleUnauthorized(t *testing.T) {
	t.Parallel()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.Logger.EXPECT().Warnf(gomock.Any(), "foo")
	response := httptest.NewRecorder()

	// Act.
	HandleUnauthorized(response, "foo", mock.Logger)

	// Assert.
	require.Equal(t, http.StatusUnauthorized, response.Code)
}
