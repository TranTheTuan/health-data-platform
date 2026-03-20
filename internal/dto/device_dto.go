package dto

// RegisterDeviceRequest is sent from the client to register a new device.
type RegisterDeviceRequest struct {
	IMEI string `json:"imei" form:"imei"`
	Name string `json:"name" form:"name"`
}

// DeviceResponse represents the device details returned to clients.
type DeviceResponse struct {
	ID         string  `json:"id"`
	IMEI       string  `json:"imei"`
	Name       string  `json:"name"`
	LastSeenAt *string `json:"last_seen_at"` // ISO8601 formatted string or nil
	CreatedAt  string  `json:"created_at"`
}
