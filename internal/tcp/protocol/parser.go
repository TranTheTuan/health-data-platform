// Package protocol implements frame parsing and reply building for the IW smartwatch protocol.
//
// Frame format: IW<CMD><PAYLOAD>#
//   - "IW" is a fixed 2-char prefix
//   - <CMD> is a 4-char command code (e.g. "AP00", "APHT")
//   - <PAYLOAD> is everything between CMD and the '#' terminator
//   - '#' is the frame terminator (NOT a newline)
//
// Devices may send fragmented TCP segments. Use ScanFrame with bufio.Scanner
// to accumulate bytes until '#' is received.
package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Command codes for all 13 IW protocol packet types.
const (
	CmdLogin       = "AP00" // IMEI login/auth      → IWBP00;<yyyyMMddHHmmss>,<tz>#
	CmdGPSLoc      = "AP01" // GPS+LBS location     → IWBP01#
	CmdLBSLoc      = "AP02" // Multi-base LBS loc   → IWBP02#
	CmdHeartbeat   = "AP03" // Keepalive ping       → IWBP03#
	CmdAudio       = "AP07" // Audio upload         → IWBP07#
	CmdAlarm       = "AP10" // Alarm + return addr  → IWBP10#
	CmdHeartRate   = "AP49" // Heart rate           → IWBP49#
	CmdHRAndBP     = "APHT" // HR + blood pressure  → IWBPHT#
	CmdHRBPSPO2    = "APHP" // HR+BP+SPO2+glucose   → IWBPHP#
	CmdTemperature = "AP50" // Body temperature     → IWBP50#
	CmdSleep       = "AP97" // Sleep data           → IWBP97#
	CmdWeather     = "APWT" // Weather sync request → IWBPWT#
	CmdECG         = "APHD" // ECG upload           → IWBPHD#
)

// replyMap maps each inbound command to its outbound BP reply prefix (without '#').
// The '#' terminator is appended by BuildReply.
var replyMap = map[string]string{
	CmdLogin:       "IWBP00", // special: timestamp appended separately
	CmdGPSLoc:      "IWBP01",
	CmdLBSLoc:      "IWBP02",
	CmdHeartbeat:   "IWBP03",
	CmdAudio:       "IWBP07",
	CmdAlarm:       "IWBP10",
	CmdHeartRate:   "IWBP49",
	CmdHRAndBP:     "IWBPHT",
	CmdHRBPSPO2:    "IWBPHP",
	CmdTemperature: "IWBP50",
	CmdSleep:       "IWBP97",
	CmdWeather:     "IWBPWT",
	CmdECG:         "IWBPHD",
}

// Sentinel errors.
var (
	ErrInvalidFrame = errors.New("protocol: missing IW prefix")
	ErrTooShort     = errors.New("protocol: payload is too short")
	ErrUnknownCmd   = errors.New("protocol: unknown command code")
	ErrInvalidIMEI  = errors.New("protocol: IMEI must be exactly 15 decimal digits")
)

// reIMEI validates that an IMEI is exactly 15 ASCII decimal digits.
var reIMEI = regexp.MustCompile(`^\d{15}$`)

// Frame holds a parsed IW protocol frame.
type Frame struct {
	Cmd     string // 4-char command code, e.g. "AP00"
	Payload string // everything after CMD and before '#'
}

// ScanFrame is a bufio.SplitFunc that splits on the '#' frame terminator.
// It handles fragmented TCP reads correctly — tokens are only returned once
// a complete frame ending with '#' has been accumulated in the buffer.
//
// Anti-DoS: callers should set bufio.Scanner.Buffer to max 64KB.
func ScanFrame(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '#'); i >= 0 {
		// Return everything up to (but not including) '#', advance past '#'
		return i + 1, data[:i], nil
	}

	// No '#' yet — request more data
	if atEOF {
		return 0, nil, nil
	}
	return 0, nil, nil
}

// ParseFrame parses a raw IW frame string (with '#' already stripped by ScanFrame).
// It validates the "IW" prefix, extracts the 4-char command code, and returns the payload.
func ParseFrame(raw string) (Frame, error) {
	// Minimum frame: "IW" + 4-char CMD = 6 characters
	if len(raw) < 6 {
		return Frame{}, ErrTooShort
	}
	if raw[:2] != "IW" {
		return Frame{}, ErrInvalidFrame
	}

	cmd := raw[2:6]
	payload := raw[6:]

	return Frame{Cmd: cmd, Payload: payload}, nil
}

// ParseAP00 extracts and validates the IMEI from an AP00 login payload.
// The payload is the raw IMEI string (15 decimal digits).
func ParseAP00(payload string) (string, error) {
	// Trim any trailing whitespace that might be present in malformed packets
	imei := payload
	if len(imei) != 15 || !reIMEI.MatchString(imei) {
		return "", fmt.Errorf("%w: got %q", ErrInvalidIMEI, imei)
	}
	return imei, nil
}

// BuildReply constructs the correct server acknowledgment string for a given command.
//
// For AP00 (login), the reply includes the current UTC timestamp and timezone offset:
//
//	IWBP00;20260319103000,0#
//
// For all other commands, the reply is simply:
//
//	IW<BPxx>#
func BuildReply(cmd string) string {
	prefix, ok := replyMap[cmd]
	if !ok {
		// Unknown command — return a generic error reply; caller should close connection
		return "IWBPERR#"
	}

	if cmd == CmdLogin {
		ts := time.Now().UTC().Format("20060102150405")
		// Timezone offset 0 = UTC; smartwatch uses this to display local time
		return fmt.Sprintf("%s;%s,0#", prefix, ts)
	}

	return prefix + "#"
}
