// Package protocol implements frame parsing and reply building for the Wonlex/Setracker smartwatch protocol.
//
// Frame format: [manufacturer*deviceID*contentLength*content]
//   - manufacturer: "3G" for device frames, "CS" for server replies
//   - deviceID: 10-digit decimal device identifier
//   - contentLength: 4-char uppercase hex (e.g. "00BC" = 188 bytes)
//   - content: CMD alone or CMD,payload (comma-separated)
//
// Devices may send fragmented TCP segments. Use ScanFrame with bufio.Scanner
// to accumulate bytes until ']' is received.
package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Command codes for Wonlex protocol packet types.
const (
	CmdLink     = "LK"     // Keep-alive/link     → [CS*deviceID*0002*LK]
	CmdLocation = "UD"     // GPS+LBS+WiFi report → no reply
	CmdBlind    = "UD2"    // Blind spot fill-in   → no reply
	CmdAlarm    = "AL"     // Alarm data report    → [3G*deviceID*0002*AL]
	CmdConfig   = "CONFIG" // Device config report → [CS*deviceID*0008*CONFIG,1]
)

// replyRule defines how to reply to a given command.
type replyRule struct {
	prefix     string // manufacturer prefix in reply ("CS" or "3G")
	needsReply bool
}

var replyRules = map[string]replyRule{
	CmdLink:     {"CS", true},
	CmdLocation: {"", false},
	CmdBlind:    {"", false},
	CmdAlarm:    {"3G", true},
	CmdConfig:   {"CS", true},
}

// Sentinel errors.
var (
	ErrInvalidFrame    = errors.New("protocol: invalid frame format")
	ErrInvalidDeviceID = errors.New("protocol: device ID must be exactly 10 decimal digits")
)

// Frame holds a parsed Wonlex protocol frame.
type Frame struct {
	Manufacturer string // "3G" from device frames
	DeviceID     string // 10-digit device identifier
	Length       string // 4-char hex content length
	Cmd          string // command code: "UD", "AL", "LK", "UD2", "CONFIG"
	Payload      string // everything after CMD comma, or "" if no payload
}

// ScanFrame is a bufio.SplitFunc that splits on the Wonlex frame delimiters '[' and ']'.
// It handles fragmented TCP reads correctly — tokens are only returned once
// a complete frame enclosed in '[...]' has been accumulated in the buffer.
//
// Anti-DoS: callers should set bufio.Scanner.Buffer to max 64KB.
func ScanFrame(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := bytes.IndexByte(data, '[')
	if start < 0 {
		return 0, nil, nil
	}

	end := bytes.IndexByte(data[start:], ']')
	if end < 0 {
		if atEOF {
			return 0, nil, nil
		}
		return 0, nil, nil
	}

	// Return content between '[' and ']', advance past ']'
	return start + end + 1, data[start+1 : start+end], nil
}

// ParseFrame parses a raw Wonlex frame string (with '[' and ']' already stripped by ScanFrame).
// It splits by '*' into 4 fields and extracts CMD and payload from content.
func ParseFrame(raw string) (Frame, error) {
	parts := strings.SplitN(raw, "*", 4)
	if len(parts) != 4 {
		return Frame{}, ErrInvalidFrame
	}

	manufacturer := parts[0]
	deviceID := parts[1]
	length := parts[2]
	content := parts[3]

	if len(deviceID) != 10 {
		return Frame{}, ErrInvalidDeviceID
	}

	// Split content into CMD and payload at first comma
	cmd, payload := content, ""
	if idx := strings.IndexByte(content, ','); idx >= 0 {
		cmd = content[:idx]
		payload = content[idx+1:]
	}

	return Frame{
		Manufacturer: manufacturer,
		DeviceID:     deviceID,
		Length:       length,
		Cmd:          cmd,
		Payload:      payload,
	}, nil
}

// BuildReply constructs the server acknowledgment string for a given command and device ID.
// Returns "" for commands that need no reply (UD, UD2).
func BuildReply(deviceID, cmd string) string {
	rule, ok := replyRules[cmd]
	if !ok || !rule.needsReply {
		return ""
	}

	content := cmd
	if cmd == CmdConfig {
		content = "CONFIG,1" // acknowledge with result=1 (OK)
	}

	lenHex := fmt.Sprintf("%04X", len(content))
	return fmt.Sprintf("[%s*%s*%s*%s]", rule.prefix, deviceID, lenHex, content)
}
