package parameters

import (
	"context"
	"fmt"

	"github.com/aleffnull/shortener/internal/pkg/database"
)

type AppParameters interface {
	Init(context.Context) error
	GetJWTSingningKey() string
}

type appParametersImpl struct {
	connection     database.Connection
	jwtSingningKey string
}

var _ AppParameters = (*appParametersImpl)(nil)

func NewAppParameters(connection database.Connection) AppParameters {
	return &appParametersImpl{
		connection: connection,
	}
}

func (i *appParametersImpl) Init(ctx context.Context) error {
	var jwtSingningKey string
	err := i.connection.QueryRow(
		ctx,
		&jwtSingningKey,
		"select value_str from app_parameters where id = 'jwt_signing_key'")
	if err != nil {
		return fmt.Errorf("appParametersImpl.Init, connection.QueryRow failed: %w", err)
	}

	i.jwtSingningKey = jwtSingningKey
	return nil
}

func (i *appParametersImpl) GetJWTSingningKey() string {
	return i.jwtSingningKey
}
