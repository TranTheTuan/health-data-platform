package tcp

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"github.com/TranTheTuan/health-data-platform/internal/dto"
	"github.com/TranTheTuan/health-data-platform/internal/service"
	"github.com/TranTheTuan/health-data-platform/internal/tcp/protocol"
)

const (
	authTimeout   = 10 * time.Second
	idleTimeout   = 5 * time.Minute
	maxFrameBytes = 64 * 1024
)

// TCPConnectHandler translates raw TCP input streams into service calls.
type TCPConnectHandler struct {
	svc service.DeviceService
}

func NewTCPConnectHandler(svc service.DeviceService) *TCPConnectHandler {
	return &TCPConnectHandler{svc: svc}
}

// HandleConnection processes the connection state loop.
func (h *TCPConnectHandler) HandleConnection(conn net.Conn) {
	defer conn.Close()
	ctx := context.Background()

	scanner := bufio.NewScanner(conn)
	scanner.Split(protocol.ScanFrame)
	buf := make([]byte, maxFrameBytes)
	scanner.Buffer(buf, maxFrameBytes)

	if err := conn.SetDeadline(time.Now().Add(authTimeout)); err != nil {
		return
	}

	if !scanner.Scan() {
		return
	}

	rawFrame := scanner.Text()
	frame, err := protocol.ParseFrame(rawFrame)
	if err != nil || frame.Cmd != protocol.CmdLogin {
		return
	}

	imei, err := protocol.ParseAP00(frame.Payload)
	if err != nil {
		return
	}

	// Talk to service using Domain directly here is permitted if retrieving state.
	// But according to strict plan, we should use Domain.
	device, err := h.svc.LookupDeviceByIMEI(ctx, imei)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return
		}
		log.Printf("tcp: db lookup error for IMEI: %v", err)
		return
	}

	if err := h.svc.UpdateLastSeen(ctx, device.ID); err != nil {
		log.Printf("tcp: update last_seen failed for device %s: %v", device.ID, err)
	}

	reply := protocol.BuildReply(frame.Cmd)
	if _, err := conn.Write([]byte(reply)); err != nil {
		return
	}

	if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
		return
	}

	for scanner.Scan() {
		if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
			return
		}

		rawFrame := strings.TrimSpace(scanner.Text())
		frame, err := protocol.ParseFrame(rawFrame)
		if err != nil {
			continue
		}

		if shouldPersist(frame.Cmd) {
			ingestReq := dto.IngestPacketRequest{
				DeviceID:    device.ID,
				UserID:      device.UserID,
				CommandCode: frame.Cmd,
				RawPayload:  frame.Payload,
				ParsedData:  nil, // Assuming generic raw payload for now
			}
			if err := h.svc.ProcessPacket(ctx, ingestReq); err != nil {
				log.Printf("tcp: insert packet error device %s cmd %s: %v", device.ID, frame.Cmd, err)
			}
		}

		replyStr := protocol.BuildReply(frame.Cmd)
		if _, err := conn.Write([]byte(replyStr)); err != nil {
			return
		}
	}
}

func shouldPersist(cmd string) bool {
	switch cmd {
	case protocol.CmdHeartbeat, protocol.CmdWeather:
		return false
	default:
		return true
	}
}
