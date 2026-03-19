// Package device provides the repository layer for device management.
// Devices are linked by IMEI to a user's Google OAuth account.
package device

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrDuplicateIMEI is returned when an IMEI is already registered (by any user).
var ErrDuplicateIMEI = errors.New("device: IMEI already registered")

// DeviceRow holds all fields returned from the devices table.
type DeviceRow struct {
	ID          string
	IMEI        string
	UserID      string
	Name        string
	LastSeenAt  *time.Time // nullable
	CreatedAt   time.Time
}

// RegisterDevice inserts a new device linking imei to userID.
// Returns ErrDuplicateIMEI if the IMEI is already taken (by any user — 409 in handler).
func RegisterDevice(ctx context.Context, db *sql.DB, userID, imei, name string) (DeviceRow, error) {
	const q = `
		INSERT INTO devices (imei, user_id, name)
		VALUES ($1, $2, $3)
		RETURNING id, imei, user_id, name, last_seen_at, created_at
	`
	row := db.QueryRowContext(ctx, q, imei, userID, name)
	return scanDeviceRow(row)
}

// ListDevices returns all devices registered to the given userID.
func ListDevices(ctx context.Context, db *sql.DB, userID string) ([]DeviceRow, error) {
	const q = `
		SELECT id, imei, user_id, name, last_seen_at, created_at
		FROM devices
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("device: list: %w", err)
	}
	defer rows.Close()

	var result []DeviceRow
	for rows.Next() {
		row, err := scanDeviceRowFromRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// scanDeviceRow scans a *sql.Row (single row) into a DeviceRow.
func scanDeviceRow(row *sql.Row) (DeviceRow, error) {
	var d DeviceRow
	var name sql.NullString
	var lastSeen sql.NullTime

	if err := row.Scan(&d.ID, &d.IMEI, &d.UserID, &name, &lastSeen, &d.CreatedAt); err != nil {
		if isDuplicateError(err) {
			return DeviceRow{}, ErrDuplicateIMEI
		}
		return DeviceRow{}, fmt.Errorf("device: scan: %w", err)
	}

	d.Name = name.String
	if lastSeen.Valid {
		d.LastSeenAt = &lastSeen.Time
	}
	return d, nil
}

// scanDeviceRowFromRows scans a *sql.Rows into a DeviceRow.
func scanDeviceRowFromRows(rows *sql.Rows) (DeviceRow, error) {
	var d DeviceRow
	var name sql.NullString
	var lastSeen sql.NullTime

	if err := rows.Scan(&d.ID, &d.IMEI, &d.UserID, &name, &lastSeen, &d.CreatedAt); err != nil {
		return DeviceRow{}, fmt.Errorf("device: scan row: %w", err)
	}

	d.Name = name.String
	if lastSeen.Valid {
		d.LastSeenAt = &lastSeen.Time
	}
	return d, nil
}

// isDuplicateError detects PostgreSQL unique constraint violations.
// pgx surfaces these as pq-style errors or as strings containing "duplicate key".
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	// pgx v5: error message contains "duplicate key value violates unique constraint"
	return containsSubstr(err.Error(), "duplicate key")
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsAny(s, sub))
}

func containsAny(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
