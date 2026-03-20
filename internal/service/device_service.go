package service

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"time"

	"github.com/TranTheTuan/health-data-platform/internal/domain"
	"github.com/TranTheTuan/health-data-platform/internal/dto"
	"github.com/TranTheTuan/health-data-platform/internal/repository"
)

var (
	ErrInvalidIMEI   = errors.New("imei must be exactly 15 decimal digits")
	ErrDuplicateIMEI = repository.ErrDuplicateIMEI
	ErrNotFound      = repository.ErrNotFound
)

var reIMEI = regexp.MustCompile(`^\d{15}$`)

// DeviceService encapsulates business logic related to device registration and streaming.
type DeviceService interface {
	RegisterDevice(ctx context.Context, userID string, req dto.RegisterDeviceRequest) (dto.DeviceResponse, error)
	ListDevices(ctx context.Context, userID string) ([]dto.DeviceResponse, error)
	LookupDeviceByIMEI(ctx context.Context, imei string) (domain.Device, error)
	UpdateLastSeen(ctx context.Context, deviceID string) error
	ProcessPacket(ctx context.Context, req dto.IngestPacketRequest) error
}

type deviceService struct {
	repo repository.DeviceRepository
}

func NewDeviceService(repo repository.DeviceRepository) DeviceService {
	return &deviceService{repo: repo}
}

func (s *deviceService) RegisterDevice(ctx context.Context, userID string, req dto.RegisterDeviceRequest) (dto.DeviceResponse, error) {
	if !reIMEI.MatchString(req.IMEI) {
		slog.Error("Invalid IMEI format attempt", slog.String("imei", req.IMEI), slog.String("user_id", userID))
		return dto.DeviceResponse{}, ErrInvalidIMEI
	}

	d, err := s.repo.RegisterDevice(ctx, userID, req.IMEI, req.Name)
	if err != nil {
		slog.Error("Device registration failed", slog.String("imei", req.IMEI), slog.Any("error", err))
		return dto.DeviceResponse{}, err
	}

	return toDeviceResponse(d), nil
}

func (s *deviceService) ListDevices(ctx context.Context, userID string) ([]dto.DeviceResponse, error) {
	devices, err := s.repo.ListDevices(ctx, userID)
	if err != nil {
		slog.Error("List devices failed", slog.String("user_id", userID), slog.Any("error", err))
		return nil, err
	}

	resp := make([]dto.DeviceResponse, 0, len(devices))
	for _, d := range devices {
		resp = append(resp, toDeviceResponse(d))
	}
	return resp, nil
}

func (s *deviceService) LookupDeviceByIMEI(ctx context.Context, imei string) (domain.Device, error) {
	// TCP handler needs the domain model to persist connection state.
	d, err := s.repo.LookupDeviceByIMEI(ctx, imei)
	if err != nil {
		slog.Error("Lookup device failed", slog.String("imei", imei), slog.Any("error", err))
	}
	return d, err
}

func (s *deviceService) UpdateLastSeen(ctx context.Context, deviceID string) error {
	err := s.repo.UpdateLastSeen(ctx, deviceID)
	if err != nil {
		slog.Error("Update last seen failed", slog.String("device_id", deviceID), slog.Any("error", err))
	}
	return err
}

func (s *deviceService) ProcessPacket(ctx context.Context, req dto.IngestPacketRequest) error {
	pkt := domain.Packet{
		DeviceID:    req.DeviceID,
		UserID:      req.UserID,
		CommandCode: req.CommandCode,
		RawPayload:  req.RawPayload,
		ParsedData:  req.ParsedData,
	}
	err := s.repo.InsertPacket(ctx, pkt)
	if err != nil {
		slog.Error("Insert packet failed", slog.String("device_id", req.DeviceID), slog.Any("error", err))
	}
	return err
}

func toDeviceResponse(d domain.Device) dto.DeviceResponse {
	resp := dto.DeviceResponse{
		ID:        d.ID,
		IMEI:      d.IMEI,
		Name:      d.Name,
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
	}
	if d.LastSeenAt != nil {
		s := d.LastSeenAt.Format(time.RFC3339)
		resp.LastSeenAt = &s
	}
	return resp
}
