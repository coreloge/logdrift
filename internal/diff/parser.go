package diff

import (
	"encoding/json"
	"fmt"
	"time"
)

// jsonLog is a minimal representation of a JSON-structured log line.
type jsonLog struct {
	Level   string `json:"level"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

// ParseLine attempts to parse a raw log line into an Entry for the given service.
// It supports JSON-structured logs; unrecognised lines are stored as-is with
// level "unknown" and the raw text as the message.
func ParseLine(service, raw string) Entry {
	entry := Entry{
		Service:   service,
		Timestamp: time.Now(),
		Raw:       raw,
	}

	var jl jsonLog
	if err := json.Unmarshal([]byte(raw), &jl); err != nil {
		// Fallback: treat whole line as plain-text message.
		entry.Level = "unknown"
		entry.Message = raw
		return entry
	}

	entry.Level = jl.Level
	if jl.Msg != "" {
		entry.Message = jl.Msg
	} else {
		entry.Message = jl.Message
	}

	if jl.Time != "" {
		if t, err := time.Parse(time.RFC3339, jl.Time); err == nil {
			entry.Timestamp = t
		}
	}

	return entry
}

// FormatEntry returns a display string for an entry suitable for terminal output.
func FormatEntry(e Entry) string {
	return fmt.Sprintf("[%s] %-7s %s", e.Service, e.Level, e.Message)
}
