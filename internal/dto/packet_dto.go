package dto

import "encoding/json"

// IngestPacketRequest is the payload coming from the TCP handler representing a parsed protocol frame.
type IngestPacketRequest struct {
	DeviceID    string
	UserID      string
	CommandCode string
	RawPayload  string
	ParsedData  json.RawMessage
}
