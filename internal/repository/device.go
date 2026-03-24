package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/TranTheTuan/health-data-platform/internal/domain"
)

var (
	ErrDuplicateIMEI = errors.New("repository: IMEI already registered")
	ErrNotFound      = errors.New("repository: not found")
)

type DeviceRepository interface {
	RegisterDevice(ctx context.Context, userID, imei, name string) (domain.Device, error)
	ListDevices(ctx context.Context, userID string) ([]domain.Device, error)
	LookupDeviceByIMEI(ctx context.Context, imei string) (domain.Device, error)
	UpdateLastSeen(ctx context.Context, deviceID string) error
	InsertPacket(ctx context.Context, pkt domain.Packet) error
	ListPackets(ctx context.Context, userID, deviceID string, packetType *string, from, to *time.Time, limit, offset int) ([]domain.Packet, int, error)
}

type pgDeviceRepo struct {
	db *sql.DB
}

func NewDeviceRepository(db *sql.DB) DeviceRepository {
	return &pgDeviceRepo{db: db}
}

func (r *pgDeviceRepo) RegisterDevice(ctx context.Context, userID, imei, name string) (domain.Device, error) {
	const q = `
		INSERT INTO devices (imei, user_id, name)
		VALUES ($1, $2, $3)
		RETURNING id, imei, user_id, name, last_seen_at, created_at
	`
	row := r.db.QueryRowContext(ctx, q, imei, userID, name)
	return scanDevice(row)
}

func (r *pgDeviceRepo) ListDevices(ctx context.Context, userID string) ([]domain.Device, error) {
	const q = `
		SELECT id, imei, user_id, name, last_seen_at, created_at
		FROM devices
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("device repo list: %w", err)
	}
	defer rows.Close()

	var result []domain.Device
	for rows.Next() {
		d, err := scanDeviceFromRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func (r *pgDeviceRepo) LookupDeviceByIMEI(ctx context.Context, imei string) (domain.Device, error) {
	// Match by last 10 digits to support both 10-digit Wonlex device IDs and legacy 15-digit IMEIs
	const q = `SELECT id, imei, user_id, name, last_seen_at, created_at FROM devices WHERE RIGHT(imei, 10) = $1 LIMIT 1`

	row := r.db.QueryRowContext(ctx, q, imei)
	d, err := scanDevice(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrNotFound) {
			return domain.Device{}, ErrNotFound
		}
		return domain.Device{}, fmt.Errorf("device repo lookup: %w", err)
	}
	return d, nil
}

func (r *pgDeviceRepo) UpdateLastSeen(ctx context.Context, deviceID string) error {
	const q = `UPDATE devices SET last_seen_at = NOW() WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, q, deviceID); err != nil {
		return fmt.Errorf("device repo update last_seen: %w", err)
	}
	return nil
}

func (r *pgDeviceRepo) InsertPacket(ctx context.Context, pkt domain.Packet) error {
	const q = `
		INSERT INTO device_packets (device_id, user_id, packet_type, raw_payload, parsed_data)
		VALUES ($1, $2, $3, $4, $5::jsonb)
	`
	var parsed interface{}
	if pkt.ParsedData != nil {
		parsed = string(pkt.ParsedData)
	}
	if _, err := r.db.ExecContext(ctx, q, pkt.DeviceID, pkt.UserID, pkt.CommandCode, pkt.RawPayload, parsed); err != nil {
		return fmt.Errorf("device repo insert packet: %w", err)
	}
	return nil
}

func (r *pgDeviceRepo) ListPackets(ctx context.Context, userID, deviceID string, packetType *string, from, to *time.Time, limit, offset int) ([]domain.Packet, int, error) {
	queryCount := `SELECT count(*) FROM device_packets WHERE user_id = $1 AND device_id = $2`
	queryPackets := `SELECT id, device_id, user_id, packet_type, raw_payload, parsed_data, recorded_at
                     FROM device_packets WHERE user_id = $1 AND device_id = $2`

	args := []interface{}{userID, deviceID}
	argID := 3

	if packetType != nil && *packetType != "" {
		condition := fmt.Sprintf(" AND packet_type = $%d", argID)
		queryCount += condition
		queryPackets += condition
		args = append(args, *packetType)
		argID++
	}
	if from != nil {
		condition := fmt.Sprintf(" AND recorded_at >= $%d", argID)
		queryCount += condition
		queryPackets += condition
		args = append(args, *from)
		argID++
	}
	if to != nil {
		condition := fmt.Sprintf(" AND recorded_at <= $%d", argID)
		queryCount += condition
		queryPackets += condition
		args = append(args, *to)
		argID++
	}

	queryPackets += fmt.Sprintf(" ORDER BY recorded_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)

	// Execute Count
	var total int
	err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("device repo count packets: %w", err)
	}

	if total == 0 {
		return []domain.Packet{}, 0, nil
	}

	// Execute Fetch
	argsFetch := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, queryPackets, argsFetch...)
	if err != nil {
		return nil, 0, fmt.Errorf("device repo fetch packets: %w", err)
	}
	defer rows.Close()

	var result []domain.Packet
	for rows.Next() {
		var p domain.Packet
		var parsed []byte
		if err := rows.Scan(&p.ID, &p.DeviceID, &p.UserID, &p.CommandCode, &p.RawPayload, &parsed, &p.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("device repo scan packet: %w", err)
		}
		if parsed != nil {
			p.ParsedData = parsed
		}
		result = append(result, p)
	}

	return result, total, rows.Err()
}

// Helpers

func scanDevice(row *sql.Row) (domain.Device, error) {
	var d domain.Device
	var name sql.NullString
	var lastSeen sql.NullTime

	if err := row.Scan(&d.ID, &d.IMEI, &d.UserID, &name, &lastSeen, &d.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Device{}, ErrNotFound
		}
		if isDuplicateError(err) {
			return domain.Device{}, ErrDuplicateIMEI
		}
		return domain.Device{}, fmt.Errorf("scan: %w", err)
	}

	d.Name = name.String
	if lastSeen.Valid {
		d.LastSeenAt = &lastSeen.Time
	}
	return d, nil
}

func scanDeviceFromRows(rows *sql.Rows) (domain.Device, error) {
	var d domain.Device
	var name sql.NullString
	var lastSeen sql.NullTime

	if err := rows.Scan(&d.ID, &d.IMEI, &d.UserID, &name, &lastSeen, &d.CreatedAt); err != nil {
		return domain.Device{}, fmt.Errorf("scan row: %w", err)
	}

	d.Name = name.String
	if lastSeen.Valid {
		d.LastSeenAt = &lastSeen.Time
	}
	return d, nil
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
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
