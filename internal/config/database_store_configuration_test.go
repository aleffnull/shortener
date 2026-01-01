package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseStoreConfiguration_String(t *testing.T) {
	t.Parallel()

	// Arrange.
	configuration := DatabaseStoreConfiguration{
		KeyStoreConfiguration: KeyStoreConfiguration{
			KeyLength:        1,
			KeyMaxLength:     2,
			KeyMaxIterations: 3,
		},
		DataSourceName: "somedb://somewhere",
	}

	// Act.
	str := configuration.String()

	// Assert.
	require.NotEmpty(t, str)
}

func TestDatabaseStoreConfiguration_NewDatabaseStoreConfiguration(t *testing.T) {
	t.Parallel()

	// Act
	configuration := NewDatabaseStoreConfiguration("somedb://somewhere")

	// Assert.
	require.Greater(t, configuration.KeyLength, 0)
	require.Greater(t, configuration.KeyMaxLength, 0)
	require.Greater(t, configuration.KeyMaxIterations, 0)
	require.Equal(t, "somedb://somewhere", configuration.DataSourceName)
}

func TestDatabaseStoreConfiguration_IsDatabaseEnabled(t *testing.T) {
	t.Parallel()

	type args struct {
		dataSourceName string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "WHEN no data source THEN disabled",
			args: args{
				dataSourceName: "",
			},
			want: false,
		},
		{
			name: "WHEN has data source THEN enabled",
			args: args{
				dataSourceName: "somedb://somewhere",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			configuration := NewDatabaseStoreConfiguration(tt.args.dataSourceName)

			// Assert.
			require.Equal(t, tt.want, configuration.IsDatabaseEnabled())
		})
	}
}
