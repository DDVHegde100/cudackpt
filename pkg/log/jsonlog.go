package log

import (
	"encoding/json"
	"os"
	"time"
)

type Event struct {
	Time  string         `json:"time"`
	Event string         `json:"event"`
	Level string         `json:"level"`
	Error string         `json:"error,omitempty"`
	Fields map[string]any `json:"fields,omitempty"`
}

func Info(name string, fields map[string]any) {
	emit("info", name, nil, fields)
}

func Error(name string, err error, fields map[string]any) {
	emit("error", name, err, fields)
}

func emit(level, name string, err error, fields map[string]any) {
	e := Event{
		Time:   time.Now().UTC().Format(time.RFC3339Nano),
		Event:  name,
		Level:  level,
		Fields: fields,
	}
	if err != nil {
		e.Error = err.Error()
	}
	b, _ := json.Marshal(e)
	out := os.Stderr
	if p := os.Getenv("CUDACKPT_LOG"); p != "" {
		if f, oerr := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); oerr == nil {
			out = f
			defer func() { _ = f.Close() }()
		}
	}
	_, _ = out.Write(append(b, '\n'))
}
