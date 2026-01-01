package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserIDHandler(t *testing.T) {
	t.Parallel()

	const token = "token"

	type args struct {
		options UserIDOptions
	}

	tests := []struct {
		name       string
		args       args
		hookBefore func(mock *mocks.Mock) *http.Request
		hookAfter  func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "WHEN get user id error THEN internal error",
			hookBefore: func(mock *mocks.Mock) *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				request.AddCookie(&http.Cookie{Name: userIDCookieName, Value: token})
				mock.AuthorizationService.EXPECT().GetUserIDFromToken(token).Return(uuid.UUID{}, assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return request
			},
			hookAfter: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "GIVEN valid token required WHEN invalid token THEN unauthorized",
			args: args{
				options: UserIDOptionsRequireValidToken,
			},
			hookBefore: func(mock *mocks.Mock) *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				request.AddCookie(&http.Cookie{Name: userIDCookieName, Value: token})
				mock.AuthorizationService.EXPECT().GetUserIDFromToken(token).Return(uuid.UUID{}, service.ErrTokenExpired)
				mock.Logger.EXPECT().Warnf(gomock.Any(), gomock.Any())
				return request
			},
			hookAfter: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "WHEN no token and set user id error THEN internal error",
			hookBefore: func(mock *mocks.Mock) *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				mock.AuthorizationService.EXPECT().CreateToken(gomock.Any()).Return("", assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return request
			},
			hookAfter: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "WHEN no token and no errors THEN set token",
			hookBefore: func(mock *mocks.Mock) *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				mock.AuthorizationService.EXPECT().CreateToken(gomock.Any()).Return(token, nil)
				return request
			},
			hookAfter: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				result := recorder.Result()
				defer result.Body.Close()

				require.Equal(t, 1, len(result.Cookies()))
				cookie := result.Cookies()[0]
				require.Equal(t, userIDCookieName, cookie.Name)
				require.Equal(t, token, cookie.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			userIDHandler := UserIDHandler(
				func(w http.ResponseWriter, r *http.Request) {
					id, ok := r.Context().Value(userIDContextKey).(uuid.UUID)
					require.True(t, ok)
					require.NotEmpty(t, id)

					w.WriteHeader(http.StatusOK)
				},
				mock.AuthorizationService,
				mock.Logger,
				tt.args.options)

			recorder := httptest.NewRecorder()
			request := tt.hookBefore(mock)

			// Act.
			userIDHandler(recorder, request)

			// Assert.
			tt.hookAfter(recorder)
		})
	}
}

func TestUserIDHandler_GetUserIDFromContext(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}

	id := uuid.New()

	tests := []struct {
		name string
		args args
		want uuid.UUID
	}{
		{
			name: "WHEN no value in context THEN empty uuid",
			args: args{
				ctx: context.Background(),
			},
		},
		{
			name: "WHEN uuid in context THEN uuid returned",
			args: args{
				ctx: context.WithValue(context.Background(), userIDContextKey, id),
			},
			want: id,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act-assert.
			id := GetUserIDFromContext(tt.args.ctx)
			require.Equal(t, tt.want, id)
		})
	}
}

func TestUserIDHandler_getUserID(t *testing.T) {
	t.Parallel()

	const token = "token"
	id := uuid.New()

	type want struct {
		id          uuid.UUID
		tokenStatus tokenStatus
		wantError   bool
	}

	tests := []struct {
		name       string
		want       *want
		hookBefore func(mock *mocks.Mock) *http.Request
	}{
		{
			name: "WHEN no cookie THEN empty token",
			want: &want{
				tokenStatus: tokenStatusEmpty,
			},
			hookBefore: func(_ *mocks.Mock) *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/foo", nil)
			},
		},
		{
			name: "WHEN token expired THEN invalid token",
			want: &want{
				tokenStatus: tokenStatusInvalid,
			},
			hookBefore: func(mock *mocks.Mock) *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				request.AddCookie(&http.Cookie{Name: userIDCookieName, Value: token})
				mock.AuthorizationService.EXPECT().GetUserIDFromToken(token).Return(uuid.UUID{}, service.ErrTokenExpired)
				return request
			},
		},
		{
			name: "WHEN token invalid THEN invalid token",
			want: &want{
				tokenStatus: tokenStatusInvalid,
			},
			hookBefore: func(mock *mocks.Mock) *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				request.AddCookie(&http.Cookie{Name: userIDCookieName, Value: token})
				mock.AuthorizationService.EXPECT().GetUserIDFromToken(token).Return(uuid.UUID{}, service.ErrTokenInvalid)
				return request
			},
		},
		{
			name: "WHEN service error THEN error",
			want: &want{
				tokenStatus: tokenStatusUnknown,
				wantError:   true,
			},
			hookBefore: func(mock *mocks.Mock) *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				request.AddCookie(&http.Cookie{Name: userIDCookieName, Value: token})
				mock.AuthorizationService.EXPECT().GetUserIDFromToken(token).Return(uuid.UUID{}, assert.AnError)
				return request
			},
		},
		{
			name: "WHEN no errors THEN valid token",
			want: &want{
				id:          id,
				tokenStatus: tokenStatusValid,
			},
			hookBefore: func(mock *mocks.Mock) *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				request.AddCookie(&http.Cookie{Name: userIDCookieName, Value: token})
				mock.AuthorizationService.EXPECT().GetUserIDFromToken(token).Return(id, nil)
				return request
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			request := tt.hookBefore(mock)

			// Act.
			id, tokenStatus, err := getUserID(request, mock.AuthorizationService)

			// Assert.
			require.Equal(t, tt.want.id, id)
			require.Equal(t, tt.want.tokenStatus, tokenStatus)
			if tt.want.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserIDHandler_setUserID(t *testing.T) {
	t.Parallel()

	const token = "token"
	id := uuid.New()

	tests := []struct {
		name       string
		wantError  bool
		hookBefore func(mock *mocks.Mock)
		hookAfter  func(response *httptest.ResponseRecorder)
	}{
		{
			name:      "WHEN service error THEN error",
			wantError: true,
			hookBefore: func(mock *mocks.Mock) {
				mock.AuthorizationService.EXPECT().CreateToken(id).Return("", assert.AnError)
			},
		},
		{
			name: "WHEN no errors THEN set cookie",
			hookBefore: func(mock *mocks.Mock) {
				mock.AuthorizationService.EXPECT().CreateToken(id).Return(token, nil)
			},
			hookAfter: func(response *httptest.ResponseRecorder) {
				result := response.Result()
				defer result.Body.Close()

				require.Equal(t, 1, len(result.Cookies()))
				cookie := result.Cookies()[0]
				require.Equal(t, userIDCookieName, cookie.Name)
				require.Equal(t, token, cookie.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock)
			response := httptest.NewRecorder()

			// Act.
			err := setUserID(id, response, mock.AuthorizationService)

			// Assert.
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.hookAfter != nil {
				tt.hookAfter(response)
			}
		})
	}
}
