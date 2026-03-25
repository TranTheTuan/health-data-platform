package protocol

import (
	"testing"
)

// TestScanFrame verifies '[...]' splitting behavior.
func TestScanFrame(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		atEOF     bool
		wantAdv   int
		wantToken string
		wantNil   bool // expect nil token (no complete frame)
	}{
		{
			name:      "complete frame",
			data:      "[3G*8800000015*0002*LK]",
			wantAdv:   23,
			wantToken: "3G*8800000015*0002*LK",
		},
		{
			name:      "frame with leading garbage",
			data:      "garbage[3G*8800000015*0002*LK]",
			wantAdv:   30,
			wantToken: "3G*8800000015*0002*LK",
		},
		{
			name:    "partial frame no closing bracket",
			data:    "[3G*8800000015",
			atEOF:   false,
			wantNil: true,
		},
		{
			name:    "empty data",
			data:    "",
			wantNil: true,
		},
		{
			name:    "no opening bracket",
			data:    "random data",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adv, tok, err := ScanFrame([]byte(tc.data), tc.atEOF)
			if err != nil {
				t.Fatalf("ScanFrame error: %v", err)
			}
			if tc.wantNil {
				if tok != nil || adv != 0 {
					t.Errorf("expected nil token and 0 advance, got adv=%d tok=%q", adv, tok)
				}
				return
			}
			if adv != tc.wantAdv {
				t.Errorf("advance = %d, want %d", adv, tc.wantAdv)
			}
			if string(tok) != tc.wantToken {
				t.Errorf("token = %q, want %q", string(tok), tc.wantToken)
			}
		})
	}
}

// TestParseFrame covers valid and invalid Wonlex frame inputs.
func TestParseFrame(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		wantManuf    string
		wantDeviceID string
		wantLength   string
		wantCmd      string
		wantPayload  string
		wantErr      bool
	}{
		{
			name:         "LK keep-alive frame",
			raw:          "3G*8800000015*0002*LK",
			wantManuf:    "3G",
			wantDeviceID: "8800000015",
			wantLength:   "0002",
			wantCmd:      "LK",
			wantPayload:  "",
		},
		{
			name:         "UD location frame with payload",
			raw:          "3G*8800000015*00BC*UD,120118,150405,A,22.570720,N,113.862017,E,0.00,0.0,0.0,6",
			wantManuf:    "3G",
			wantDeviceID: "8800000015",
			wantLength:   "00BC",
			wantCmd:      "UD",
			wantPayload:  "120118,150405,A,22.570720,N,113.862017,E,0.00,0.0,0.0,6",
		},
		{
			name:         "AL alarm frame with payload",
			raw:          "3G*8800000015*0030*AL,120118,150405,A,22.5,N,113.8,E,0.0,180.0",
			wantManuf:    "3G",
			wantDeviceID: "8800000015",
			wantLength:   "0030",
			wantCmd:      "AL",
			wantPayload:  "120118,150405,A,22.5,N,113.8,E,0.0,180.0",
		},
		{
			name:         "CONFIG frame",
			raw:          "3G*8800000015*0006*CONFIG",
			wantManuf:    "3G",
			wantDeviceID: "8800000015",
			wantLength:   "0006",
			wantCmd:      "CONFIG",
			wantPayload:  "",
		},
		{
			name:    "too few fields (missing content)",
			raw:     "3G*8800000015*0002",
			wantErr: true,
		},
		{
			name:    "device ID not 10 digits",
			raw:     "3G*123456789*0002*LK",
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
			if got.Manufacturer != tc.wantManuf {
				t.Errorf("Manufacturer = %q, want %q", got.Manufacturer, tc.wantManuf)
			}
			if got.DeviceID != tc.wantDeviceID {
				t.Errorf("DeviceID = %q, want %q", got.DeviceID, tc.wantDeviceID)
			}
			if got.Length != tc.wantLength {
				t.Errorf("Length = %q, want %q", got.Length, tc.wantLength)
			}
			if got.Cmd != tc.wantCmd {
				t.Errorf("Cmd = %q, want %q", got.Cmd, tc.wantCmd)
			}
			if got.Payload != tc.wantPayload {
				t.Errorf("Payload = %q, want %q", got.Payload, tc.wantPayload)
			}
		})
	}
}

// TestBuildReply verifies reply strings for all Wonlex commands.
func TestBuildReply(t *testing.T) {
	const (
		deviceID = "8800000015"
		manuf3G  = "3G"
		manufXY  = "XY" // test dynamic prefix
	)

	tests := []struct {
		manuf string
		cmd   string
		want  string
	}{
		{manuf3G, CmdLink, "[3G*8800000015*0002*LK]"},
		{manuf3G, CmdAlarm, "[3G*8800000015*0002*AL]"},
		{manuf3G, CmdConfig, "[3G*8800000015*0008*CONFIG,1]"},
		{manufXY, CmdLink, "[XY*8800000015*0002*LK]"},
		{manuf3G, CmdLocation, ""}, // no reply
		{manuf3G, CmdBlind, ""},    // no reply
		{manuf3G, "UNKNOWN", ""},   // unknown cmd → no reply
	}

	for _, tc := range tests {
		t.Run("BuildReply_"+tc.cmd+"_"+tc.manuf, func(t *testing.T) {
			got := BuildReply(tc.manuf, deviceID, tc.cmd)
			if got != tc.want {
				t.Errorf("BuildReply(%q, %q, %q) = %q, want %q", tc.manuf, deviceID, tc.cmd, got, tc.want)
			}
		})
	}
}
