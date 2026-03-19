-- Migration: 001_initial_schema.sql
-- Creates devices and device_packets tables for smartwatch TCP ingestion

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- devices: maps IMEI (hardware identity) to a user account
CREATE TABLE IF NOT EXISTS devices (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    imei         VARCHAR(15)  NOT NULL UNIQUE,
    user_id      TEXT         NOT NULL,
    name         TEXT,
    last_seen_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- device_packets: single table for all 13 IW protocol packet types
-- parsed_data JSONB allows flexible per-type field storage without per-type tables (YAGNI)
CREATE TABLE IF NOT EXISTS device_packets (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id   UUID         NOT NULL REFERENCES devices(id),
    user_id     TEXT         NOT NULL,  -- denormalized for fast per-user queries without JOIN
    packet_type VARCHAR(10)  NOT NULL,
    raw_payload TEXT         NOT NULL,
    parsed_data JSONB,
    recorded_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index: recent health data per device + type
CREATE INDEX IF NOT EXISTS idx_device_packets_device_type_time
    ON device_packets(device_id, packet_type, recorded_at DESC);

-- Index: all health data for a specific user
CREATE INDEX IF NOT EXISTS idx_device_packets_user_time
    ON device_packets(user_id, recorded_at DESC);
