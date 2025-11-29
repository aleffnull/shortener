package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/models"
)

func TestShortenerApp_GetURL(t *testing.T) {
	const (
		key      = "foo"
		shortURL = "http://localhost/bar"
	)

	// Arrange.
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.Store.EXPECT().Load(gomock.Any(), key).Return(&domain.URLItem{URL: shortURL}, nil)
	configuration := &config.Configuration{}
	shortener := NewShortenerApp(mock.Connection, mock.Store, mock.Logger, mock.AppParameters, configuration)

	// Act.
	item, err := shortener.GetURL(ctx, key)

	// Assert.
	require.NotNil(t, item)
	require.Equal(t, shortURL, item.URL)
	require.NoError(t, err)
}

func TestShortenerApp_ShortenURL(t *testing.T) {
	tests := []struct {
		name          string
		configuration *config.Configuration
		request       *models.ShortenRequest
		response      *models.ShortenResponse
		wantError     bool
		hookBefore    func(request *models.ShortenRequest, mock *mocks.Mock)
	}{
		{
			name:          "storage error",
			configuration: &config.Configuration{},
			request: &models.ShortenRequest{
				URL: "foo",
			},
			wantError: true,
			hookBefore: func(request *models.ShortenRequest, mock *mocks.Mock) {
				mock.Store.EXPECT().Save(gomock.Any(), request.URL, gomock.Any()).Return("", assert.AnError)
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
			hookBefore: func(request *models.ShortenRequest, mock *mocks.Mock) {
				mock.Store.EXPECT().Save(gomock.Any(), request.URL, gomock.Any()).Return("key", nil)
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
			hookBefore: func(request *models.ShortenRequest, mock *mocks.Mock) {
				mock.Store.EXPECT().Save(gomock.Any(), request.URL, gomock.Any()).Return("key", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(tt.request, mock)
			shortener := NewShortenerApp(mock.Connection, mock.Store, mock.Logger, mock.AppParameters, tt.configuration)

			response, err := shortener.ShortenURL(ctx, tt.request, uuid.New())
			require.Equal(t, tt.response, response)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
