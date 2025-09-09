package authorization

import (
	"errors"
	"fmt"
	"time"

	"github.com/aleffnull/shortener/internal/pkg/parameters"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

type Service interface {
	CreateToken(uuid.UUID) (string, error)
	GetUserIDFromToken(string) (uuid.UUID, error)
}

type serviceImpl struct {
	parameters parameters.AppParameters
}

var _ Service = (*serviceImpl)(nil)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

const jwtTokenDuration = 24 * time.Hour

func NewAuthorizationService(parameters parameters.AppParameters) Service {
	return &serviceImpl{
		parameters: parameters,
	}
}

func (i *serviceImpl) CreateToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTokenDuration)),
			},
			UserID: userID,
		})
	tokenString, err := token.SignedString([]byte(i.parameters.GetJWTSigningKey()))
	if err != nil {
		return "", fmt.Errorf("authorizationServiceImpl.CreateToken, token.SignedString failed: %w", err)
	}

	return tokenString, nil
}

func (i *serviceImpl) GetUserIDFromToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			return []byte(i.parameters.GetJWTSigningKey()), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.UUID{}, ErrTokenExpired
		}

		return uuid.UUID{}, fmt.Errorf("authorizationServiceImpl.GetUserIDFromToken, jwt.ParseWithClaims failed: %w", err)
	}

	if !token.Valid {
		return uuid.UUID{}, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("authorizationServiceImpl.GetUserIDFromToken, token.Claims.(*Claims) failed: %w", err)
	}

	return claims.UserID, nil
}
