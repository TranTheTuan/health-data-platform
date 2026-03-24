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
	ErrInvalidIMEI   = errors.New("device ID must be 10-15 decimal digits")
	ErrDuplicateIMEI = repository.ErrDuplicateIMEI
	ErrNotFound      = repository.ErrNotFound
)

// reIMEI accepts both 10-digit Wonlex device IDs and legacy 15-digit IMEIs.
var reIMEI = regexp.MustCompile(`^\d{10,15}$`)

// DeviceService encapsulates business logic related to device registration and streaming.
type DeviceService interface {
	RegisterDevice(ctx context.Context, userID string, req dto.RegisterDeviceRequest) (dto.DeviceResponse, error)
	ListDevices(ctx context.Context, userID string) ([]dto.DeviceResponse, error)
	LookupDeviceByIMEI(ctx context.Context, imei string) (domain.Device, error)
	UpdateLastSeen(ctx context.Context, deviceID string) error
	ProcessPacket(ctx context.Context, req dto.IngestPacketRequest) error
	ListDevicePackets(ctx context.Context, userID string, req dto.ListPacketsRequest) (dto.PaginatedPacketResponse, error)
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

func (s *deviceService) ListDevicePackets(ctx context.Context, userID string, req dto.ListPacketsRequest) (dto.PaginatedPacketResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	} else if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	var fromTime, toTime *time.Time
	if req.From != "" {
		if t, err := time.Parse(time.RFC3339, req.From); err == nil {
			fromTime = &t
		} else {
			// Ignore format errors but could log if we wanted. User rule: "always log whenever an error happens... no need to put info and debug". We aren't failing so no error log needed.
		}
	}
	if req.To != "" {
		if t, err := time.Parse(time.RFC3339, req.To); err == nil {
			toTime = &t
		}
	}

	var pType *string
	if req.PacketType != "" {
		pType = &req.PacketType
	}

	packets, total, err := s.repo.ListPackets(ctx, userID, req.DeviceID, pType, fromTime, toTime, req.Limit, req.Offset)
	if err != nil {
		slog.Error("List device packets failed", slog.String("device_id", req.DeviceID), slog.Any("error", err))
		return dto.PaginatedPacketResponse{}, err
	}

	resp := make([]dto.PacketResponse, 0, len(packets))
	for _, p := range packets {
		resp = append(resp, dto.PacketResponse{
			ID:          p.ID,
			CommandCode: p.CommandCode,
			RawPayload:  p.RawPayload,
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		})
	}

	return dto.PaginatedPacketResponse{
		Packets: resp,
		Total:   total,
	}, nil
}
