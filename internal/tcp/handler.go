// Package tcp implements the per-connection handler for the IW smartwatch TCP protocol.
// Each accepted connection gets its own goroutine running HandleConnection.
package tcp

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"github.com/TranTheTuan/health-data-platform/internal/tcp/protocol"
)

const (
	// authTimeout is how long we wait for the initial AP00 login frame.
	// Prevents slow-open / slow-login attacks.
	authTimeout = 10 * time.Second

	// idleTimeout is reset after each valid data frame. Kills zombie connections.
	idleTimeout = 5 * time.Minute

	// maxFrameBytes is the scanner buffer limit to prevent DoS via unbounded frames.
	maxFrameBytes = 64 * 1024 // 64 KB
)

// HandleConnection manages the full lifecycle of one smartwatch connection:
//  1. Auth: first frame MUST be AP00 with valid IMEI.
//  2. Data loop: parse frame → store in DB → reply.
//
// On any error or auth failure the connection is silently closed.
func HandleConnection(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	ctx := context.Background()

	scanner := bufio.NewScanner(conn)
	scanner.Split(protocol.ScanFrame)

	// Anti-DoS: limit per-frame buffer to 64 KB
	buf := make([]byte, maxFrameBytes)
	scanner.Buffer(buf, maxFrameBytes)

	// ── Step 1: Auth — first frame must be AP00 ──────────────────────────────
	if err := conn.SetDeadline(time.Now().Add(authTimeout)); err != nil {
		return
	}

	if !scanner.Scan() {
		// Connection closed or error before first frame
		return
	}

	rawFrame := scanner.Text()
	frame, err := protocol.ParseFrame(rawFrame)
	if err != nil || frame.Cmd != protocol.CmdLogin {
		// First frame is not a valid AP00 — reject silently
		return
	}

	imei, err := protocol.ParseAP00(frame.Payload)
	if err != nil {
		// Invalid IMEI format — reject silently
		return
	}

	device, err := LookupDeviceByIMEI(ctx, db, imei)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Unknown device — silent drop (no error response that leaks info)
			return
		}
		log.Printf("tcp: db lookup error for IMEI (redacted): %v", err)
		return
	}

	// Update last_seen_at; non-fatal if it fails
	if err := UpdateLastSeen(ctx, db, device.ID); err != nil {
		log.Printf("tcp: update last_seen failed for device %s: %v", device.ID, err)
	}

	// Send AP00 reply with UTC timestamp
	reply := protocol.BuildReply(frame.Cmd)
	if _, err := conn.Write([]byte(reply)); err != nil {
		return
	}

	// ── Step 2: Data loop ────────────────────────────────────────────────────
	// Reset idle timeout for the data phase; it resets on every valid frame.
	if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
		return
	}

	for scanner.Scan() {
		// Reset idle deadline on each received frame
		if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
			return
		}

		rawFrame := scanner.Text()
		rawFrame = strings.TrimSpace(rawFrame)
		frame, err := protocol.ParseFrame(rawFrame)
		if err != nil {
			// Skip malformed frames but keep connection alive
			log.Printf("tcp: malformed frame from device %s: %v", device.ID, err)
			continue
		}

		// Persist packet to DB — heartbeat and weather replies require no storage
		if shouldPersist(frame.Cmd) {
			if err := InsertPacket(ctx, db, device.ID, device.UserID, frame.Cmd, frame.Payload, nil); err != nil {
				log.Printf("tcp: insert packet error device %s cmd %s: %v", device.ID, frame.Cmd, err)
				// Non-fatal: still send reply so the watch doesn't hang
			}
		}

		// Always reply — watches hang or retry if no ack arrives
		replyStr := protocol.BuildReply(frame.Cmd)
		if _, err := conn.Write([]byte(replyStr)); err != nil {
			return
		}
	}

	if err := scanner.Err(); err != nil {
		// Normal idle timeout or remote close logs at debug level only
		log.Printf("tcp: scanner closed for device %s: %v", device.ID, err)
	}
}

// shouldPersist returns false for packet types that need no DB storage:
//   - AP03 (heartbeat): purely a keepalive
//   - APWT (weather): device request, server replies with weather data (out of scope)
func shouldPersist(cmd string) bool {
	switch cmd {
	case protocol.CmdHeartbeat, protocol.CmdWeather:
		return false
	default:
		return true
	}
}
