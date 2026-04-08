package audit

import (
	"encoding/json"
	"os"
	"sync"
)

// FileListener writes audit events to a file, one JSON per line.
type FileListener struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileListener creates a new FileListener for the given file path.
// The file is kept open for the lifetime of the listener.
func NewFileListener(path string) (*FileListener, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileListener{file: file}, nil
}

// OnEvent appends the event as a JSON line to the file.
func (f *FileListener) OnEvent(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	f.mu.Lock()
	_, err = f.file.Write(data)
	f.mu.Unlock()
	return err
}

// Close closes the underlying file.
func (f *FileListener) Close() error {
	return f.file.Close()
}
