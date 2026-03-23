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

type ListPacketsRequest struct {
	DeviceID   string `query:"-"`
	PacketType string `query:"type"`
	From       string `query:"from"`
	To         string `query:"to"`
	Limit      int    `query:"limit"`
	Offset     int    `query:"offset"`
}

type PacketResponse struct {
	ID          string `json:"id"`
	CommandCode string `json:"command_code"`
	RawPayload  string `json:"raw_payload"`
	CreatedAt   string `json:"created_at"`
}

type PaginatedPacketResponse struct {
	Packets []PacketResponse `json:"packets"`
	Total   int              `json:"total"`
}
