package store

import (
	"os"
	"path"
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestFileStore_LoadAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		want       []*domain.ColdStoreEntry
		wantError  bool
		hookBefore func() *config.Configuration
	}{
		{
			name: "WHEN file not exists THEN no entries",
			hookBefore: func() *config.Configuration {
				filePath := path.Join(t.TempDir(), "store.jsonl")
				return &config.Configuration{
					FileStore: &config.FileStoreConfiguration{
						FilePath: filePath,
					},
				}
			},
		},
		{
			name:      "WHEN file can't be opened THEN error",
			wantError: true,
			hookBefore: func() *config.Configuration {
				filePath := path.Join(t.TempDir(), "store.jsonl")
				require.NoError(t, os.WriteFile(filePath, []byte{}, 0222))
				return &config.Configuration{
					FileStore: &config.FileStoreConfiguration{
						FilePath: filePath,
					},
				}
			},
		},
		{
			name:      "WHEN unmarshal error THEN error",
			wantError: true,
			hookBefore: func() *config.Configuration {
				filePath := path.Join(t.TempDir(), "store.jsonl")
				require.NoError(t, os.WriteFile(filePath, []byte("foo"), 0644))
				return &config.Configuration{
					FileStore: &config.FileStoreConfiguration{
						FilePath: filePath,
					},
				}
			},
		},
		{
			name: "WHEN no errors THEN ok",
			want: []*domain.ColdStoreEntry{
				{
					Key:   "foo",
					Value: "http://bar.buz",
				},
			},
			hookBefore: func() *config.Configuration {
				filePath := path.Join(t.TempDir(), "store.jsonl")
				require.NoError(t, os.WriteFile(filePath, []byte(`{"key":"foo","value":"http://bar.buz"}`), 0644))
				return &config.Configuration{
					FileStore: &config.FileStoreConfiguration{
						FilePath: filePath,
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			store := NewFileStore(tt.hookBefore())

			// Act.
			entries, err := store.LoadAll()

			// Assert.
			require.ElementsMatch(t, tt.want, entries)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFileStore_Save(t *testing.T) {
	t.Parallel()

	// Arrange.
	filePath := path.Join(t.TempDir(), "store.jsonl")
	configuration := &config.Configuration{
		FileStore: &config.FileStoreConfiguration{
			FilePath: filePath,
		},
	}
	store := NewFileStore(configuration)

	// Act.
	err := store.Save(&domain.ColdStoreEntry{
		Key:   "foo",
		Value: "http://bar.buz",
	})

	// Assert
	require.NoError(t, err)
	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)
	require.Greater(t, fileInfo.Size(), int64(0))
}
