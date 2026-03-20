package protocol

import (
	"strings"
	"testing"
)

// TestParseFrame covers valid and invalid frame inputs.
func TestParseFrame(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantCmd string
		wantPay string
		wantErr bool
	}{
		{
			name:    "valid AP00 login frame",
			raw:     "IWAP00353456789012345",
			wantCmd: "AP00",
			wantPay: "353456789012345",
		},
		{
			name:    "valid APHT health frame",
			raw:     "IWAPHT72,80,120",
			wantCmd: "APHT",
			wantPay: "72,80,120",
		},
		{
			name:    "valid AP03 heartbeat (empty payload)",
			raw:     "IWAP03",
			wantCmd: "AP03",
			wantPay: "",
		},
		{
			name:    "valid APHP frame",
			raw:     "IWAPHPpayloaddata",
			wantCmd: "APHP",
			wantPay: "payloaddata",
		},
		{
			name:    "missing IW prefix",
			raw:     "AP00353456789012345",
			wantErr: true,
		},
		{
			name:    "frame too short (only IW)",
			raw:     "IW",
			wantErr: true,
		},
		{
			name:    "frame too short (IW + 3 chars)",
			raw:     "IWAP0",
			wantErr: true,
		},
		{
			name:    "empty string",
			raw:     "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseFrame(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ParseFrame(%q) = nil error, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFrame(%q) error: %v", tc.raw, err)
			}
			if got.Cmd != tc.wantCmd {
				t.Errorf("Cmd = %q, want %q", got.Cmd, tc.wantCmd)
			}
			if got.Payload != tc.wantPay {
				t.Errorf("Payload = %q, want %q", got.Payload, tc.wantPay)
			}
		})
	}
}

// TestParseAP00 covers valid and invalid IMEI strings.
func TestParseAP00(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantIMEI string
		wantErr  bool
	}{
		{
			name:     "valid 15-digit IMEI",
			payload:  "353456789012345",
			wantIMEI: "353456789012345",
		},
		{
			name:    "non-digit characters",
			payload: "35345678901234A",
			wantErr: true,
		},
		{
			name:    "14-digit IMEI (too short)",
			payload: "35345678901234",
			wantErr: true,
		},
		{
			name:    "16-digit IMEI (too long)",
			payload: "3534567890123456",
			wantErr: true,
		},
		{
			name:    "empty string",
			payload: "",
			wantErr: true,
		},
		{
			name:    "IMEI with spaces",
			payload: "353456 89012345",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseAP00(tc.payload)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ParseAP00(%q) = nil error, want error", tc.payload)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseAP00(%q) error: %v", tc.payload, err)
			}
			if got != tc.wantIMEI {
				t.Errorf("IMEI = %q, want %q", got, tc.wantIMEI)
			}
		})
	}
}

// TestBuildReply verifies reply strings for all known commands.
func TestBuildReply(t *testing.T) {
	tests := []struct {
		cmd        string
		wantPrefix string
		wantSuffix string
	}{
		{CmdLogin, "IWBP00;", "#"},
		{CmdGPSLoc, "IWBP01", "#"},
		{CmdLBSLoc, "IWBP02", "#"},
		{CmdHeartbeat, "IWBP03", "#"},
		{CmdAudio, "IWBP07", "#"},
		{CmdAlarm, "IWBP10", "#"},
		{CmdHeartRate, "IWBP49", "#"},
		{CmdHRAndBP, "IWBPHT", "#"},
		{CmdHRBPSPO2, "IWBPHP", "#"},
		{CmdTemperature, "IWBP50", "#"},
		{CmdSleep, "IWBP97", "#"},
		{CmdWeather, "IWBPWT", "#"},
		{CmdECG, "IWBPHD", "#"},
	}

	for _, tc := range tests {
		t.Run("BuildReply_"+tc.cmd, func(t *testing.T) {
			reply := BuildReply(tc.cmd)
			if !strings.HasPrefix(reply, tc.wantPrefix) {
				t.Errorf("BuildReply(%q) = %q, want prefix %q", tc.cmd, reply, tc.wantPrefix)
			}
			if !strings.HasSuffix(reply, tc.wantSuffix) {
				t.Errorf("BuildReply(%q) = %q, want suffix %q", tc.cmd, reply, tc.wantSuffix)
			}
		})
	}
}

// TestBuildReplyAP00Format validates the AP00 reply contains a well-formed timestamp.
func TestBuildReplyAP00Format(t *testing.T) {
	reply := BuildReply(CmdLogin)
	// Expected format: IWBP00;YYYYMMDDHHmmss,0#
	// e.g. IWBP00;20260319103000,0#
	if !strings.HasPrefix(reply, "IWBP00;") {
		t.Errorf("AP00 reply %q must start with IWBP00;", reply)
	}
	if !strings.HasSuffix(reply, "#") {
		t.Errorf("AP00 reply %q must end with #", reply)
	}

	// Strip IWBP00; prefix and # suffix, should be "YYYYMMDDHHmmss,0"
	inner := strings.TrimPrefix(reply, "IWBP00;")
	inner = strings.TrimSuffix(inner, "#")
	parts := strings.Split(inner, ",")
	if len(parts) != 2 {
		t.Errorf("AP00 reply inner %q should have format <timestamp>,<tz>", inner)
	}
	if len(parts[0]) != 14 {
		t.Errorf("AP00 timestamp %q should be 14 chars (YYYYMMDDHHmmss)", parts[0])
	}
}

// TestBuildReplyUnknown verifies unknown commands return an error reply.
func TestBuildReplyUnknown(t *testing.T) {
	reply := BuildReply("APXX")
	if reply != "IWBPERR#" {
		t.Errorf("unknown cmd should return IWBPERR#, got %q", reply)
	}
}

// TestScanFrame verifies # splitting behavior.
func TestScanFrame(t *testing.T) {
	// Full frame: advance past '#', token is the frame without '#'
	data := []byte("IWAP00353456789012345#")
	adv, tok, err := ScanFrame(data, false)
	if err != nil {
		t.Fatalf("ScanFrame error: %v", err)
	}
	if adv != len(data) {
		t.Errorf("advance = %d, want %d", adv, len(data))
	}
	if string(tok) != "IWAP00353456789012345" {
		t.Errorf("token = %q, want %q", string(tok), "IWAP00353456789012345")
	}
}

// TestScanFramePartial verifies that incomplete frames (no '#') return 0, nil, nil.
func TestScanFramePartial(t *testing.T) {
	data := []byte("IWAP00353") // no '#'
	adv, tok, err := ScanFrame(data, false)
	if err != nil {
		t.Fatalf("ScanFrame partial error: %v", err)
	}
	if adv != 0 || tok != nil {
		t.Errorf("partial frame should return 0, nil; got adv=%d tok=%q", adv, tok)
	}
}
