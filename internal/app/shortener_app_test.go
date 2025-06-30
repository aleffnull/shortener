package app

import (
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/store"
	"github.com/aleffnull/shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mock struct {
	storage     *mocks.MockStore
	coldStorage *mocks.MockColdStore
	logger      *mocks.MockLogger
}

func newMock(ctrl *gomock.Controller) *mock {
	return &mock{
		storage:     mocks.NewMockStore(ctrl),
		coldStorage: mocks.NewMockColdStore(ctrl),
		logger:      mocks.NewMockLogger(ctrl),
	}
}

func TestShortenerApp_GetURL(t *testing.T) {
	const (
		key      = "foo"
		shortURL = "http://localhost/bar"
	)

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := newMock(ctrl)
	mock.storage.EXPECT().Load(key).Return(shortURL, true)
	configuration := &config.Configuration{}
	shortener := NewShortenerApp(mock.storage, mock.coldStorage, mock.logger, configuration)

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
		hookBefore    func(request *models.ShortenRequest, mock *mock)
	}{
		{
			name:          "storage error",
			configuration: &config.Configuration{},
			request: &models.ShortenRequest{
				URL: "foo",
			},
			wantError: true,
			hookBefore: func(request *models.ShortenRequest, mock *mock) {
				mock.storage.EXPECT().Save(request.URL).Return("", assert.AnError)
			},
		},
		{
			name:          "cold storage error",
			configuration: &config.Configuration{},
			request: &models.ShortenRequest{
				URL: "foo",
			},
			wantError: true,
			hookBefore: func(request *models.ShortenRequest, mock *mock) {
				coldStoreEntry := &store.ColdStoreEntry{
					Key:   "key",
					Value: request.URL,
				}
				mock.storage.EXPECT().Save(request.URL).Return("key", nil)
				mock.coldStorage.EXPECT().Save(coldStoreEntry).Return(assert.AnError)
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
			hookBefore: func(request *models.ShortenRequest, mock *mock) {
				coldStoreEntry := &store.ColdStoreEntry{
					Key:   "key",
					Value: request.URL,
				}
				mock.storage.EXPECT().Save(request.URL).Return("key", nil)
				mock.coldStorage.EXPECT().Save(coldStoreEntry).Return(nil)
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
			hookBefore: func(request *models.ShortenRequest, mock *mock) {
				coldStoreEntry := &store.ColdStoreEntry{
					Key:   "key",
					Value: request.URL,
				}
				mock.storage.EXPECT().Save(request.URL).Return("key", nil)
				mock.coldStorage.EXPECT().Save(coldStoreEntry).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mock := newMock(ctrl)
			tt.hookBefore(tt.request, mock)
			shortener := NewShortenerApp(mock.storage, mock.coldStorage, mock.logger, tt.configuration)

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
