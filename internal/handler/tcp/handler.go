package tcp

import (
	"bufio"
	"context"
	"log/slog"
	"net"
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

// HandleConnection processes the Wonlex protocol connection state loop.
// Authentication: first frame's DeviceID (10-digit) is looked up in the DB.
// Every subsequent frame is persisted; replies are sent only for commands that require them.
func (h *TCPConnectHandler) HandleConnection(conn net.Conn) {
	defer conn.Close()
	ctx := context.Background()

	scanner := bufio.NewScanner(conn)
	scanner.Split(protocol.ScanFrame)
	scanner.Buffer(make([]byte, maxFrameBytes), maxFrameBytes)

	// Auth: read first frame within authTimeout
	if err := conn.SetDeadline(time.Now().Add(authTimeout)); err != nil {
		return
	}
	if !scanner.Scan() {
		return
	}

	frame, err := protocol.ParseFrame(scanner.Text())
	if err != nil {
		return
	}

	// Device ID in frame header IS the authentication — no special login command needed
	device, err := h.svc.LookupDeviceByIMEI(ctx, frame.DeviceID)
	if err != nil {
		return
	}

	if err := h.svc.UpdateLastSeen(ctx, device.ID); err != nil {
		slog.Error("tcp: update last_seen failed", slog.String("device_id", device.ID), slog.Any("error", err))
	}

	// Reply to first frame if the command requires it
	if reply := protocol.BuildReply(frame.DeviceID, frame.Cmd); reply != "" {
		if _, err := conn.Write([]byte(reply)); err != nil {
			return
		}
	}

	h.persistPacket(ctx, device.ID, device.UserID, frame)

	// Main loop: persist all frames, reply only when needed
	if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
		return
	}
	for scanner.Scan() {
		if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
			return
		}

		frame, err := protocol.ParseFrame(scanner.Text())
		if err != nil {
			continue
		}

		h.persistPacket(ctx, device.ID, device.UserID, frame)

		if reply := protocol.BuildReply(frame.DeviceID, frame.Cmd); reply != "" {
			if _, err := conn.Write([]byte(reply)); err != nil {
				return
			}
		}
	}
}

func (h *TCPConnectHandler) persistPacket(ctx context.Context, deviceID, userID string, frame protocol.Frame) {
	req := dto.IngestPacketRequest{
		DeviceID:    deviceID,
		UserID:      userID,
		CommandCode: frame.Cmd,
		RawPayload:  frame.Payload,
	}
	if err := h.svc.ProcessPacket(ctx, req); err != nil {
		slog.Error("tcp: insert packet failed",
			slog.String("device_id", deviceID),
			slog.String("cmd", frame.Cmd),
			slog.Any("error", err),
		)
	}
}
