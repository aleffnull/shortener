package app

import (
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
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
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStore(ctrl)
	storage.EXPECT().Load(key).Return(shortURL, true)
	configuration := &config.Configuration{}
	shortener := NewShortenerApp(storage, configuration)

	// Act.
	url, ok := shortener.GetURL(key)

	// Assert.
	require.Equal(t, shortURL, url)
	require.True(t, ok)
}

func TestShortenerApp_ShortenURL(t *testing.T) {
	tests := []struct {
		name          string
		configuration *config.Configuration
		request       *models.ShortenRequest
		response      *models.ShortenResponse
		wantError     bool
		hookBefore    func(request *models.ShortenRequest, storage *mocks.MockStore)
	}{
		{
			name:          "storage error",
			configuration: &config.Configuration{},
			request: &models.ShortenRequest{
				URL: "foo",
			},
			wantError: true,
			hookBefore: func(request *models.ShortenRequest, storage *mocks.MockStore) {
				storage.EXPECT().Save(request.URL).Return("", assert.AnError)
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
			hookBefore: func(request *models.ShortenRequest, storage *mocks.MockStore) {
				storage.EXPECT().Save(request.URL).Return("key", nil)
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
			hookBefore: func(request *models.ShortenRequest, storage *mocks.MockStore) {
				storage.EXPECT().Save(request.URL).Return("key", nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockStore(ctrl)
			tt.hookBefore(tt.request, storage)
			shortener := NewShortenerApp(storage, tt.configuration)

			response, err := shortener.ShortenURL(tt.request)
			require.Equal(t, tt.response, response)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
