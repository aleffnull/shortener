package parameters

import (
	"context"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAppParameters_Init(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		want       string
		wantError  bool
		hookBefore func(mock *mocks.Mock)
	}{
		{
			name:      "WHEN connection error THEN error",
			wantError: true,
			hookBefore: func(mock *mocks.Mock) {
				mock.Connection.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
		},
		{
			name: "WHEN no errors THEN ok",
			want: "foo",
			hookBefore: func(mock *mocks.Mock) {
				mock.Connection.EXPECT().
					QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, result *string, sql string, args ...any) error {
						*result = "foo"
						return nil
					})
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
			parameters := NewAppParameters(mock.Connection)

			// Act.
			err := parameters.Init(context.Background())

			// Assert
			require.Equal(t, tt.want, parameters.GetJWTSigningKey())
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
