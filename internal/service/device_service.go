package service

import (
	"context"
	"errors"
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
		return dto.DeviceResponse{}, ErrInvalidIMEI
	}

	d, err := s.repo.RegisterDevice(ctx, userID, req.IMEI, req.Name)
	if err != nil {
		return dto.DeviceResponse{}, err
	}

	return toDeviceResponse(d), nil
}

func (s *deviceService) ListDevices(ctx context.Context, userID string) ([]dto.DeviceResponse, error) {
	devices, err := s.repo.ListDevices(ctx, userID)
	if err != nil {
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
	return s.repo.LookupDeviceByIMEI(ctx, imei)
}

func (s *deviceService) UpdateLastSeen(ctx context.Context, deviceID string) error {
	return s.repo.UpdateLastSeen(ctx, deviceID)
}

func (s *deviceService) ProcessPacket(ctx context.Context, req dto.IngestPacketRequest) error {
	pkt := domain.Packet{
		DeviceID:    req.DeviceID,
		UserID:      req.UserID,
		CommandCode: req.CommandCode,
		RawPayload:  req.RawPayload,
		ParsedData:  req.ParsedData,
	}
	return s.repo.InsertPacket(ctx, pkt)
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
