package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/parameters"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserIDOptions int

const (
	UserIDOptionsNone UserIDOptions = iota
	UserIDOptionsRequireValidToken
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

type tokenStatus int

const (
	tokenStatusUnknown tokenStatus = iota
	tokenStatusEmpty
	tokenStatusInvalid
	tokenStatusValid
)

type contextKey int

const (
	jwtTokenDuration            = 24 * time.Hour
	userIDCookieName            = "X-UserID"
	userIDContextKey contextKey = iota
)

func UserIDHandler(
	handlerFunc http.HandlerFunc,
	parameters parameters.AppParameters,
	logger logger.Logger,
	options UserIDOptions,
) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		userID, status, err := getUserID(request, parameters)
		if err != nil {
			utils.HandleServerError(response, err, logger)
			return
		}

		if status != tokenStatusValid {
			// Нет валидного токена, значит, либо его нет в принципе, либо он невалиден.

			if status == tokenStatusInvalid && options == UserIDOptionsRequireValidToken {
				// Токен есть и невалиден, но хендлер требует валидный, не можем авторизовать пользователя.
				utils.HandleUnauthorized(response, "Invalid token", logger)
				return
			}

			// Либо никакого токена не было, либо он невалиден, но нам не важно.
			userID = uuid.New()
			err = setUserID(userID, response, parameters)
			if err != nil {
				utils.HandleServerError(response, err, logger)
				return
			}
		}

		ctx := context.WithValue(request.Context(), userIDContextKey, userID)
		requestWithUserID := request.WithContext(ctx)

		handlerFunc(response, requestWithUserID)
	}
}

func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	value := ctx.Value(userIDContextKey)
	if value == nil {
		return uuid.UUID{}
	}

	return (value).(uuid.UUID)
}

func getUserID(request *http.Request, parameters parameters.AppParameters) (uuid.UUID, tokenStatus, error) {
	cookie, err := request.Cookie(userIDCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return uuid.UUID{}, tokenStatusEmpty, nil
		}

		return uuid.UUID{}, tokenStatusUnknown, fmt.Errorf("getUserID, request.Cookie failed: %w", err)
	}

	token, err := jwt.ParseWithClaims(
		cookie.Value,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			return []byte(parameters.GetJWTSingningKey()), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			// Токен просрочился, поэтому невалиден.
			return uuid.UUID{}, tokenStatusInvalid, nil
		}

		// Все остальные ошибки считаем внутренней ошибкой сервера.
		return uuid.UUID{}, tokenStatusUnknown, fmt.Errorf("getUserID, jwt.Parse failed: %w", err)
	}

	if !token.Valid {
		// Токен оказался невалиден.
		return uuid.UUID{}, tokenStatusInvalid, nil
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return uuid.UUID{}, tokenStatusUnknown, fmt.Errorf("getUserID, token.Claims.(*Claims) failed: %w", err)
	}

	return claims.UserID, tokenStatusValid, nil
}

func setUserID(userID uuid.UUID, response http.ResponseWriter, parameters parameters.AppParameters) error {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTokenDuration)),
			},
			UserID: userID,
		})
	tokenString, err := token.SignedString([]byte(parameters.GetJWTSingningKey()))
	if err != nil {
		return fmt.Errorf("setUserID, token.SignedString failed: %w", err)
	}

	http.SetCookie(response, &http.Cookie{
		Name:  userIDCookieName,
		Value: tokenString,
	})

	return nil
}
