package service

import (
	"testing"
	"time"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthorizationService_CreateToken(t *testing.T) {
	t.Parallel()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.AppParameters.EXPECT().GetJWTSigningKey().Return("")
	service := NewAuthorizationService(mock.AppParameters)

	// Act.
	token, err := service.CreateToken(uuid.New())
	require.NotEmpty(t, token)
	require.NoError(t, err)
}

func TestAuthorizationService_GetUserIDFromToken(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	signingKey := "key"

	type want struct {
		id  uuid.UUID
		err error
	}

	tests := []struct {
		name       string
		want       *want
		hookBefore func(mock *mocks.Mock) string
	}{
		{
			name: "WHEN token expired THEN expired error",
			want: &want{
				id:  uuid.UUID{},
				err: ErrTokenExpired,
			},
			hookBefore: func(mock *mocks.Mock) string {
				mock.AppParameters.EXPECT().GetJWTSigningKey().Return(signingKey)
				token := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					Claims{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
						},
						UserID: id,
					})
				tokenString, err := token.SignedString([]byte(signingKey))
				require.NoError(t, err)
				return tokenString
			},
		},
		{
			name: "WHEN wrong token THEN parsing error",
			want: &want{
				id:  uuid.UUID{},
				err: jwt.ErrTokenSignatureInvalid,
			},
			hookBefore: func(mock *mocks.Mock) string {
				token := jwt.NewWithClaims(
					jwt.SigningMethodHS512,
					Claims{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
						},
						UserID: id,
					})
				tokenString, err := token.SignedString([]byte(signingKey))
				require.NoError(t, err)
				return tokenString
			},
		},
		{
			name: "WHEN valid token THEN ok",
			want: &want{
				id: id,
			},
			hookBefore: func(mock *mocks.Mock) string {
				mock.AppParameters.EXPECT().GetJWTSigningKey().Return(signingKey)
				token := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					Claims{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
						},
						UserID: id,
					})
				tokenString, err := token.SignedString([]byte(signingKey))
				require.NoError(t, err)
				return tokenString
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tokenString := tt.hookBefore(mock)
			service := NewAuthorizationService(mock.AppParameters)

			// Act.
			id, err := service.GetUserIDFromToken(tokenString)

			// Assert.
			require.Equal(t, tt.want.id, id)
			if tt.want.err == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.want.err)
			}
		})
	}
}
