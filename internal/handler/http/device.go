package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/TranTheTuan/health-data-platform/internal/dto"
	"github.com/TranTheTuan/health-data-platform/internal/service"
)

// DeviceHandler translates HTTP requests into service calls using DTOs.
type DeviceHandler struct {
	svc service.DeviceService
}

func NewDeviceHandler(svc service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

func (h *DeviceHandler) Register(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req dto.RegisterDeviceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	resp, err := h.svc.RegisterDevice(c.Request().Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidIMEI) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if errors.Is(err, service.ErrDuplicateIMEI) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "IMEI already registered"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *DeviceHandler) List(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	resp, err := h.svc.ListDevices(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, resp)
}
