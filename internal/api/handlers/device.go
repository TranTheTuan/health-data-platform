// Package handlers contains HTTP handlers for the Health Data Platform API.
package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"

	"github.com/TranTheTuan/health-data-platform/internal/device"
)

// reIMEI validates exactly 15 decimal digits.
var reIMEI = regexp.MustCompile(`^\d{15}$`)

// DeviceHandler handles device registration and listing endpoints.
type DeviceHandler struct {
	db *sql.DB
}

// NewDeviceHandler creates a DeviceHandler with the provided database pool.
func NewDeviceHandler(db *sql.DB) *DeviceHandler {
	return &DeviceHandler{db: db}
}

// registerRequest is the JSON body for POST /protected/devices.
type registerRequest struct {
	IMEI string `json:"imei"`
	Name string `json:"name"`
}

// deviceResponse is the JSON response for device endpoints.
type deviceResponse struct {
	ID          string  `json:"id"`
	IMEI        string  `json:"imei"`
	Name        string  `json:"name,omitempty"`
	LastSeenAt  *string `json:"last_seen_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

// Register handles POST /protected/devices.
// Registers a smartwatch IMEI under the authenticated user's account.
func (h *DeviceHandler) Register(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Validate IMEI: must be exactly 15 decimal digits
	if !reIMEI.MatchString(req.IMEI) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "imei must be exactly 15 decimal digits",
		})
	}

	row, err := device.RegisterDevice(c.Request().Context(), h.db, userID, req.IMEI, req.Name)
	if err != nil {
		if errors.Is(err, device.ErrDuplicateIMEI) {
			// Do not reveal which user owns the IMEI
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "IMEI already registered",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	return c.JSON(http.StatusCreated, toDeviceResponse(row))
}

// List handles GET /protected/devices.
// Returns all devices registered to the authenticated user.
func (h *DeviceHandler) List(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	rows, err := device.ListDevices(c.Request().Context(), h.db, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	resp := make([]deviceResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, toDeviceResponse(r))
	}
	return c.JSON(http.StatusOK, resp)
}

// toDeviceResponse maps a device.DeviceRow to the JSON response shape.
func toDeviceResponse(row device.DeviceRow) deviceResponse {
	resp := deviceResponse{
		ID:        row.ID,
		IMEI:      row.IMEI,
		Name:      row.Name,
		CreatedAt: row.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if row.LastSeenAt != nil {
		s := row.LastSeenAt.Format("2006-01-02T15:04:05Z")
		resp.LastSeenAt = &s
	}
	return resp
}
