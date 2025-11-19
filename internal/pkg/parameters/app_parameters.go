package parameters

import (
	"context"
	"fmt"

	"github.com/aleffnull/shortener/internal/repository"
)

type AppParameters interface {
	Init(context.Context) error
	GetJWTSigningKey() string
}

type appParametersImpl struct {
	connection    repository.Connection
	jwtSigningKey string
}

var _ AppParameters = (*appParametersImpl)(nil)

func NewAppParameters(connection repository.Connection) AppParameters {
	return &appParametersImpl{
		connection: connection,
	}
}

func (i *appParametersImpl) Init(ctx context.Context) error {
	var jwtSigningKey string
	err := i.connection.QueryRow(
		ctx,
		&jwtSigningKey,
		"select value_str from app_parameters where id = 'jwt_signing_key'")
	if err != nil {
		return fmt.Errorf("appParametersImpl.Init, connection.QueryRow failed: %w", err)
	}

	i.jwtSigningKey = jwtSigningKey
	return nil
}

func (i *appParametersImpl) GetJWTSigningKey() string {
	return i.jwtSigningKey
}
