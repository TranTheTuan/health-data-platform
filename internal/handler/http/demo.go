package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/TranTheTuan/health-data-platform/internal/service"
)

// DemoHandler exposes HTTP endpoints for the demo TCP session feature.
type DemoHandler struct {
	demoSvc service.DemoService
	devSvc  service.DeviceService
}

func NewDemoHandler(demoSvc service.DemoService, devSvc service.DeviceService) *DemoHandler {
	return &DemoHandler{demoSvc: demoSvc, devSvc: devSvc}
}

// getIMEI fetches the IMEI for a device owned by the given user.
// Returns "", false if not found or not owned.
func (h *DemoHandler) getIMEI(c echo.Context, userID, deviceID string) (string, bool) {
	devices, err := h.devSvc.ListDevices(c.Request().Context(), userID)
	if err != nil {
		return "", false
	}
	for _, d := range devices {
		if d.ID == deviceID {
			return d.IMEI, true
		}
	}
	return "", false
}

func (h *DemoHandler) StartSession(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	deviceID := c.Param("id")
	imei, ok := h.getIMEI(c, userID, deviceID)
	if !ok {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if err := h.demoSvc.StartSession(c.Request().Context(), deviceID, imei); err != nil {
		if errors.Is(err, service.ErrDemoSessionAlreadyActive) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "session already active"})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"active": true})
}

func (h *DemoHandler) StopSession(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	deviceID := c.Param("id")
	if _, ok := h.getIMEI(c, userID, deviceID); !ok {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if err := h.demoSvc.StopSession(c.Request().Context(), deviceID); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "no active session"})
	}
	return c.JSON(http.StatusOK, map[string]bool{"active": false})
}

func (h *DemoHandler) SessionStatus(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	deviceID := c.Param("id")
	if _, ok := h.getIMEI(c, userID, deviceID); !ok {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	return c.JSON(http.StatusOK, map[string]bool{"active": h.demoSvc.IsActive(deviceID)})
}

func (h *DemoHandler) SendBurst(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	deviceID := c.Param("id")
	if _, ok := h.getIMEI(c, userID, deviceID); !ok {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	const burstCount = 7
	if err := h.demoSvc.SendBurst(c.Request().Context(), deviceID, burstCount); err != nil {
		if errors.Is(err, service.ErrDemoSessionNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "no active session"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]int{"sent": burstCount})
}
