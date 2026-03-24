-- Add functional index for 10-digit device ID suffix lookup used by the Wonlex TCP handler.
-- The handler calls LookupDeviceByIMEI with a 10-digit ID and matches RIGHT(imei, 10).
CREATE INDEX IF NOT EXISTS idx_devices_imei_suffix10
    ON devices (RIGHT(imei, 10));
