package audit

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aleffnull/shortener/internal/config"
)

type FileReceiver struct {
	auditFile string
}

var _ Receiver = (*FileReceiver)(nil)

func NewFileReceiver(configuration *config.Configuration) *FileReceiver {
	return &FileReceiver{
		auditFile: configuration.AuditFile,
	}
}

func (r *FileReceiver) AddEvent(event *Event) error {
	if len(r.auditFile) == 0 {
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("FileReceiver.AddEvent, json.Marshal failed: %w", err)
	}

	file, err := os.OpenFile(r.auditFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("FileReceiver.AddEvent, os.OpenFile failed: %w", err)
	}

	defer file.Close()

	data = append(data, '\n')
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("FileReceiver.AddEvent, file.Write failed: %w", err)
	}

	return nil
}
