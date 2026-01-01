package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileStoreConfiguration_String(t *testing.T) {
	t.Parallel()

	// Arrange.
	configuration := FileStoreConfiguration{
		FilePath: "foo.jsonl",
	}

	// Act.
	str := configuration.String()

	// Assert.
	require.NotEmpty(t, str)
}
