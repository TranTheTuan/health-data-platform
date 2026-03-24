package demo

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/TranTheTuan/health-data-platform/internal/tcp/protocol"
)

// PersistableCommands lists packet types used for demo bursts.
var PersistableCommands = []string{
	protocol.CmdLocation, // UD
	protocol.CmdAlarm,    // AL
}

// BuildWonlexFrame constructs a Wonlex frame: [3G*deviceID*LEN*CMD] or [3G*deviceID*LEN*CMD,payload]
func BuildWonlexFrame(deviceID, cmd, payload string) string {
	content := cmd
	if payload != "" {
		content = cmd + "," + payload
	}
	lenHex := fmt.Sprintf("%04X", len(content))
	return fmt.Sprintf("[3G*%s*%s*%s]", deviceID, lenHex, content)
}

// LinkFrame constructs the LK keep-alive/link frame for the given device ID.
func LinkFrame(deviceID string) string {
	return BuildWonlexFrame(deviceID, protocol.CmdLink, "")
}

// RandomFrame returns a complete Wonlex frame for a random persistable command.
func RandomFrame(deviceID string) string {
	cmd := PersistableCommands[rand.Intn(len(PersistableCommands))]
	return BuildWonlexFrame(deviceID, cmd, randomPayload(cmd))
}

func randomPayload(cmd string) string {
	switch cmd {
	case protocol.CmdLocation:
		// Simplified UD location payload (Hanoi area)
		ts := time.Now().UTC().Format("060102,150405")
		lat := 21.0 + float64(rand.Intn(1000))/10000.0
		lng := 105.8 + float64(rand.Intn(10000))/10000.0
		speed := rand.Intn(80)
		sats := 6 + rand.Intn(6)
		return fmt.Sprintf("%s,A,%.6f,N,%.6f,E,%d.00,0.0,0.0,%d,100,51,14188,0,00010010,6,255,460,0,9360,5081,156,...", ts, lat, lng, speed, sats)
	case protocol.CmdAlarm:
		// Simplified AL payload (location + alarm flags)
		ts := time.Now().UTC().Format("060102,150405")
		return fmt.Sprintf("%s,A,21.027763,N,105.834160,E,0.00,188.6,0.0,9,100,51,14188,0,00200000,...", ts)
	default:
		return ""
	}
}
