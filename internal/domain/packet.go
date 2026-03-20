package domain

import (
	"encoding/json"
	"time"
)

// Packet represents a raw parsed data packet from a smartwatch.
type Packet struct {
	ID          string // UUID
	DeviceID    string
	UserID      string
	CommandCode string          // e.g. "AP01", "APHT"
	RawPayload  string          // Raw string from the watch
	ParsedData  json.RawMessage // JSONB data resulting from parsing
	CreatedAt   time.Time
}
