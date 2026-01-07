package audit

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFileReceiver_AddEvent(t *testing.T) {
	t.Parallel()

	// Arrange.
	filePath := path.Join(t.TempDir(), "audit.jsonl")
	configuration := &config.Configuration{
		AuditFile: filePath,
	}
	receiver := NewFileReceiver(configuration)

	// Act.
	err := receiver.AddEvent(&domain.AuditEvent{
		Timestamp: domain.AuditFormattedTime(time.Date(2025, 1, 2, 18, 0, 0, 0, time.UTC)),
		Action:    domain.AuditActionShorten,
		UserID:    uuid.New(),
		URL:       "http://foo.bar",
	})

	// Assert.
	require.NoError(t, err)
	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)
	require.Greater(t, fileInfo.Size(), int64(0))
}
