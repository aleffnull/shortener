package store

import (
	"context"
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_saveWithUniqueKey(t *testing.T) {
	t.Parallel()

	type args struct {
		configuration *config.KeyStoreConfiguration
		value         string
		saver         saverFunc
	}

	tests := []struct {
		name          string
		args          *args
		wantError     bool
		wantKeyLength int
	}{
		{
			name: "WHEN saver error THEN error",
			args: &args{
				configuration: &config.KeyStoreConfiguration{
					KeyLength:        1,
					KeyMaxLength:     2,
					KeyMaxIterations: 1,
				},
				value: "foo",
				saver: func(_ context.Context, _ string, _ string) (bool, error) {
					return false, assert.AnError
				},
			},
			wantError: true,
		},
		{
			name: "WHEN always exists THEN error",
			args: &args{
				configuration: &config.KeyStoreConfiguration{
					KeyLength:        1,
					KeyMaxLength:     2,
					KeyMaxIterations: 1,
				},
				value: "foo",
				saver: func(_ context.Context, _ string, _ string) (bool, error) {
					return true, nil
				},
			},
			wantError: true,
		},
		{
			name: "WHEN not existing key THEN ok",
			args: &args{
				configuration: &config.KeyStoreConfiguration{
					KeyLength:        1,
					KeyMaxLength:     2,
					KeyMaxIterations: 1,
				},
				value: "foo",
				saver: func(_ context.Context, _ string, _ string) (bool, error) {
					return false, nil
				},
			},
			wantKeyLength: 1,
		},
		{
			name: "WHEN short key exists THEN key length is doubled",
			args: &args{
				configuration: &config.KeyStoreConfiguration{
					KeyLength:        1,
					KeyMaxLength:     10,
					KeyMaxIterations: 1,
				},
				value: "foo",
				saver: func(_ context.Context, key string, _ string) (bool, error) {
					return len(key) <= 1, nil
				},
			},
			wantKeyLength: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			keyStore := &keyStore{
				configuration: tt.args.configuration,
			}

			key, err := keyStore.saveWithUniqueKey(context.Background(), tt.args.value, tt.args.saver)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, len(key), tt.wantKeyLength)
		})
	}
}

func Test_randomString(t *testing.T) {
	t.Parallel()

	str := randomString(10)
	require.Len(t, str, 10)
}
