// Package tcp implements the repository layer for the TCP server's device and packet storage.
// Uses database/sql with pgx driver (registered in the db package).
package tcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrNotFound is returned when a device lookup finds no matching row.
var ErrNotFound = errors.New("tcp/repo: device not found")

// DeviceRecord holds the minimal device fields needed by the TCP handler.
type DeviceRecord struct {
	ID     string // UUID as string
	UserID string // Google OAuth user ID (denormalized from devices table)
}

// LookupDeviceByIMEI finds a device by its IMEI.
// Returns ErrNotFound if no row exists — callers should silently close the connection.
func LookupDeviceByIMEI(ctx context.Context, db *sql.DB, imei string) (DeviceRecord, error) {
	const q = `SELECT id, user_id FROM devices WHERE imei = $1 LIMIT 1`

	row := db.QueryRowContext(ctx, q, imei)
	var rec DeviceRecord
	if err := row.Scan(&rec.ID, &rec.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DeviceRecord{}, ErrNotFound
		}
		return DeviceRecord{}, fmt.Errorf("tcp/repo: lookup device: %w", err)
	}
	return rec, nil
}

// UpdateLastSeen sets last_seen_at = NOW() for the given device ID.
// Called on every successful AP00 login.
func UpdateLastSeen(ctx context.Context, db *sql.DB, deviceID string) error {
	const q = `UPDATE devices SET last_seen_at = NOW() WHERE id = $1`
	if _, err := db.ExecContext(ctx, q, deviceID); err != nil {
		return fmt.Errorf("tcp/repo: update last_seen: %w", err)
	}
	return nil
}

// InsertPacket stores a raw + parsed device packet into device_packets.
// parsedData may be nil for packet types where field parsing is not yet implemented.
func InsertPacket(
	ctx context.Context,
	db *sql.DB,
	deviceID, userID, packetType, rawPayload string,
	parsedData interface{},
) error {
	var jsonData []byte
	var err error

	if parsedData != nil {
		jsonData, err = json.Marshal(parsedData)
		if err != nil {
			return fmt.Errorf("tcp/repo: marshal parsed_data: %w", err)
		}
	}

	const q = `
		INSERT INTO device_packets (device_id, user_id, packet_type, raw_payload, parsed_data)
		VALUES ($1, $2, $3, $4, $5::jsonb)
	`
	args := []interface{}{deviceID, userID, packetType, rawPayload, nil}
	if jsonData != nil {
		args[4] = string(jsonData)
	}

	if _, err := db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("tcp/repo: insert packet: %w", err)
	}
	return nil
}
