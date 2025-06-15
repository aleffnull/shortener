package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/models"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/require"
)

func TestHandleGetRequest(t *testing.T) {
	type want struct {
		statusCode int
		headers    map[string]string
		emptyBody  bool
	}

	tests := []struct {
		name       string
		key        string
		hookBefore func(shortener *ShortenerApp)
		want       want
	}{
		{
			name: "unknown key",
			key:  "foo",
			want: want{
				statusCode: http.StatusBadRequest,
				emptyBody:  false,
			},
		},
		{
			name: "existing key",
			key:  "foo",
			hookBefore: func(shortener *ShortenerApp) {
				shortener.SetKeyAndURL("foo", "http://bar.buz")
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				headers: map[string]string{
					headers.Location: "http://bar.buz",
				},
				emptyBody: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			recorder := httptest.NewRecorder()
			shortenerApp := NewShortenerApp(&config.Configuration{})
			handler := NewHandler(shortenerApp)
			if tt.hookBefore != nil {
				tt.hookBefore(shortenerApp)
			}

			// Act.
			handler.HandleGetRequest(recorder, tt.key)

			// Assert.
			result := recorder.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)

			if len(tt.want.headers) != 0 {
				for header, value := range tt.want.headers {
					require.Equal(t, value, result.Header.Get(header))
				}
			}

			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			if tt.want.emptyBody {
				require.Empty(t, body)
			} else {
				require.NotEmpty(t, body)
			}
		})
	}
}

func TestHandlePostRequest(t *testing.T) {
	type want struct {
		statusCode  int
		validateURL bool
	}

	tests := []struct {
		name    string
		baseURL string
		body    string
		want    want
	}{
		{
			name: "no body",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "invalid base URL",
			baseURL: ":localhost:8080",
			body:    "http://foo.bar",
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name:    "valid request",
			baseURL: "http://localhost:8080",
			body:    "http://foo.bar",
			want: want{
				statusCode:  http.StatusCreated,
				validateURL: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			configuration := &config.Configuration{
				BaseURL: tt.baseURL,
			}
			shortenerApp := NewShortenerApp(configuration)
			handler := NewHandler(shortenerApp)

			// Act.
			handler.HandlePostRequest(recorder, request)

			// Assert.
			result := recorder.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)

			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.NotEmpty(t, body)

			if tt.want.validateURL {
				_, err = url.ParseRequestURI(string(body))
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleAPIRequest(t *testing.T) {
	type want struct {
		statusCode  int
		validateURL bool
	}

	tests := []struct {
		name           string
		baseURL        string
		shortenRequest *models.ShortenRequest
		want           want
	}{
		{
			name:           "no body",
			shortenRequest: nil,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:           "invalid request",
			shortenRequest: &models.ShortenRequest{},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "invalid base URL",
			baseURL: ":localhost:8080",
			shortenRequest: &models.ShortenRequest{
				URL: "http://foo.bar",
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name:    "valid request",
			baseURL: "http://localhost:8080",
			shortenRequest: &models.ShortenRequest{
				URL: "http://foo.bar",
			},
			want: want{
				statusCode:  http.StatusCreated,
				validateURL: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var body io.Reader
			if tt.shortenRequest != nil {
				jsonRequest, err := json.Marshal(tt.shortenRequest)
				require.NoError(t, err)
				body = bytes.NewReader(jsonRequest)
			}

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			recorder := httptest.NewRecorder()
			configuration := &config.Configuration{
				BaseURL: tt.baseURL,
			}
			shortenerApp := NewShortenerApp(configuration)
			handler := NewHandler(shortenerApp)

			// Act.
			handler.HandleAPIRequest(recorder, request)

			// Assert.
			result := recorder.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.validateURL {
				defer result.Body.Close()
				var shortenResponse models.ShortenResponse
				err := json.NewDecoder(result.Body).Decode(&shortenResponse)
				require.NoError(t, err)

				_, err = url.ParseRequestURI(string(shortenResponse.Result))
				require.NoError(t, err)
			}
		})
	}
}
