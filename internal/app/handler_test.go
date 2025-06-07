package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
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
		request    string
		hookBefore func(shortener *ShortenerApp)
		want       want
	}{
		{
			name:    "no key in URL",
			request: "/",
			want: want{
				statusCode: http.StatusBadRequest,
				emptyBody:  false,
			},
		},
		{
			name:    "unknown key",
			request: "/foo",
			want: want{
				statusCode: http.StatusBadRequest,
				emptyBody:  false,
			},
		},
		{
			name:    "existing key",
			request: "/foo",
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
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			recorder := httptest.NewRecorder()
			shortener := NewShortenerApp()
			if tt.hookBefore != nil {
				tt.hookBefore(shortener)
			}

			// Act.
			HandleGetRequest(recorder, request, shortener)

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
		name        string
		contentType string
		body        string
		want        want
	}{
		{
			name:        "invalid content type",
			contentType: mimetype.TextHTML,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "empty body",
			contentType: mimetype.TextPlain,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "valid request",
			contentType: mimetype.TextPlain,
			body:        "http://foo.bar",
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
			request.Header.Set(headers.ContentType, tt.contentType)
			recorder := httptest.NewRecorder()
			shortener := NewShortenerApp()

			// Act.
			HandlePostRequest(recorder, request, shortener)

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

func TestHandleInvalidMethod(t *testing.T) {
	// Arrange.
	recorder := httptest.NewRecorder()

	// Act.
	HandleInvalidMethod(recorder)

	// Assert.
	result := recorder.Result()
	require.Equal(t, http.StatusBadRequest, result.StatusCode)

	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
}
