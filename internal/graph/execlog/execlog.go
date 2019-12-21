package execlog

import (
	"encoding/json"
	"os"
	"time"

	"github.com/herzult/porte/internal/graph"
)

type Entry struct {
	ID            string          `json:"id"`
	GraphID       string          `json:"graphId"`
	ClientName    string          `json:"clientName"`
	ClientVersion string          `json:"clientVersion"`
	StartTime     time.Time       `json:"startTime"`
	Duration      time.Duration   `json:"duration"`
	Request       *graph.Request  `json:"request"`
	Response      *graph.Response `json:"response"`
	Error         string          `json:"error"`
}

type EntryWriter interface {
	Write(*Entry) error
}

type FileEntryWriter struct {
	File *os.File
}

func (w *FileEntryWriter) Write(entry *Entry) error {
	return json.NewEncoder(w.File).Encode(entry)
}

type AMQPLogWriter struct {
	Chann
}
