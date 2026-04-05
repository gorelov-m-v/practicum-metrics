package audit

import (
	"encoding/json"
	"os"
	"sync"
)

// FileListener writes audit events to a file, one JSON per line.
type FileListener struct {
	mu   sync.Mutex
	path string
}

// NewFileListener creates a new FileListener for the given file path.
func NewFileListener(path string) *FileListener {
	return &FileListener{path: path}
}

// OnEvent appends the event as a JSON line to the file.
func (f *FileListener) OnEvent(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
