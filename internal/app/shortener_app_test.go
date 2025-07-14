package app

import (
	"context"
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/testutils"
	"github.com/aleffnull/shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestShortenerApp_GetURL(t *testing.T) {
	const (
		key      = "foo"
		shortURL = "http://localhost/bar"
	)

	// Arrange.
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mock := testutils.NewMock(ctrl)
	mock.Store.EXPECT().Load(gomock.Any(), key).Return(shortURL, true, nil)
	configuration := &config.Configuration{}
	shortener := NewShortenerApp(mock.Store, mock.Logger, configuration)

	// Act.
	url, ok, err := shortener.GetURL(ctx, key)

	// Assert.
	require.Equal(t, shortURL, url)
	require.True(t, ok)
	require.NoError(t, err)
}

func TestShortenerApp_ShortenURL(t *testing.T) {
	tests := []struct {
		name          string
		configuration *config.Configuration
		request       *models.ShortenRequest
		response      *models.ShortenResponse
		wantError     bool
		hookBefore    func(request *models.ShortenRequest, mock *testutils.Mock)
	}{
		{
			name:          "storage error",
			configuration: &config.Configuration{},
			request: &models.ShortenRequest{
				URL: "foo",
			},
			wantError: true,
			hookBefore: func(request *models.ShortenRequest, mock *testutils.Mock) {
				mock.Store.EXPECT().Save(gomock.Any(), request.URL).Return("", assert.AnError)
			},
		},
		{
			name: "base URL error",
			configuration: &config.Configuration{
				BaseURL: ":http//localhost",
			},
			request: &models.ShortenRequest{
				URL: "foo",
			},
			wantError: true,
			hookBefore: func(request *models.ShortenRequest, mock *testutils.Mock) {
				mock.Store.EXPECT().Save(gomock.Any(), request.URL).Return("key", nil)
			},
		},
		{
			name: "valid request",
			configuration: &config.Configuration{
				BaseURL: "http://localhost",
			},
			request: &models.ShortenRequest{
				URL: "foo",
			},
			response: &models.ShortenResponse{
				Result: "http://localhost/key",
			},
			hookBefore: func(request *models.ShortenRequest, mock *testutils.Mock) {
				mock.Store.EXPECT().Save(gomock.Any(), request.URL).Return("key", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			mock := testutils.NewMock(ctrl)
			tt.hookBefore(tt.request, mock)
			shortener := NewShortenerApp(mock.Store, mock.Logger, tt.configuration)

			response, err := shortener.ShortenURL(ctx, tt.request)
			require.Equal(t, tt.response, response)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
