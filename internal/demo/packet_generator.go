package demo

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/TranTheTuan/health-data-platform/internal/tcp/protocol"
)

// PersistableCommands lists packet types valid for demo bursts.
var PersistableCommands = []string{
	protocol.CmdHeartRate,   // AP49
	protocol.CmdHRAndBP,     // APHT
	protocol.CmdHRBPSPO2,    // APHP
	protocol.CmdTemperature, // AP50
	protocol.CmdGPSLoc,      // AP01
	protocol.CmdSleep,       // AP97
}

// RandomFrame returns a complete IW frame string for a random packet type.
func RandomFrame() string {
	cmd := PersistableCommands[rand.Intn(len(PersistableCommands))]
	return BuildFrame(cmd, randomPayload(cmd))
}

// BuildFrame constructs a client-side IW frame: "IW<CMD><PAYLOAD>#"
func BuildFrame(cmd, payload string) string {
	return "IW" + cmd + payload + "#"
}

// LoginFrame constructs the AP00 login frame for the given IMEI.
func LoginFrame(imei string) string {
	return BuildFrame(protocol.CmdLogin, imei)
}

func randomPayload(cmd string) string {
	switch cmd {
	case protocol.CmdHeartRate:
		return fmt.Sprintf("%d", randRange(55, 105))
	case protocol.CmdHRAndBP:
		return fmt.Sprintf("%d,%d,%d", randRange(60, 100), randRange(100, 140), randRange(60, 90))
	case protocol.CmdHRBPSPO2:
		return fmt.Sprintf("%d,%d,%d,%d,%.1f", randRange(60, 100), randRange(100, 140), randRange(60, 90), randRange(95, 100), float32(randRange(45, 65))/10.0)
	case protocol.CmdTemperature:
		return fmt.Sprintf("%d", randRange(3600, 3750))
	case protocol.CmdGPSLoc:
		lat := 21.0 + float64(rand.Intn(100))/1000.0
		lng := 105.8 + float64(rand.Intn(100))/1000.0
		ts := time.Now().UTC().Format("20060102150405")
		return fmt.Sprintf("%.4f,%.4f,10,%d,6,%s", lat, lng, rand.Intn(5), ts)
	case protocol.CmdSleep:
		return fmt.Sprintf("%d,%d,%d", randRange(60, 180), randRange(30, 120), randRange(10, 60))
	default:
		return ""
	}
}

func randRange(min, max int) int {
	return min + rand.Intn(max-min+1)
}
