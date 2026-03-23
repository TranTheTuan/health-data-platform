package service

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/TranTheTuan/health-data-platform/internal/demo"
)

var (
	ErrDemoSessionAlreadyActive = errors.New("demo session already active for this device")
	ErrDemoSessionNotFound      = errors.New("no active demo session for this device")
)

// DemoService manages demo TCP session lifecycle for testing and demo purposes.
type DemoService interface {
	StartSession(ctx context.Context, deviceID, imei string) error
	SendBurst(ctx context.Context, deviceID string, count int) error
	StopSession(ctx context.Context, deviceID string) error
	IsActive(deviceID string) bool
}

type demoSession struct {
	conn   net.Conn
	reader *bufio.Reader
}

type demoService struct {
	mu       sync.RWMutex
	sessions map[string]*demoSession // keyed by deviceID
	tcpAddr  string
}

// NewDemoService creates a DemoService that connects to the given TCP address.
func NewDemoService(tcpAddr string) DemoService {
	return &demoService{
		sessions: make(map[string]*demoSession),
		tcpAddr:  resolveTCPAddr(tcpAddr),
	}
}

// resolveTCPAddr converts listen-only addresses (":9090", "0.0.0.0:9090")
// to dialable loopback addresses ("localhost:9090").
func resolveTCPAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "localhost" + addr
	}
	if strings.HasPrefix(addr, "0.0.0.0:") {
		return "localhost:" + strings.TrimPrefix(addr, "0.0.0.0:")
	}
	return addr
}

func (s *demoService) StartSession(_ context.Context, deviceID, imei string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[deviceID]; exists {
		return ErrDemoSessionAlreadyActive
	}

	conn, err := net.DialTimeout("tcp", s.tcpAddr, 5*time.Second)
	if err != nil {
		return err
	}

	if _, err = conn.Write([]byte(demo.LoginFrame(imei))); err != nil {
		conn.Close()
		return err
	}

	conn.SetDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(conn)
	ack, err := reader.ReadString('#')
	if err != nil || !strings.HasPrefix(ack, "IWBP00") {
		conn.Close()
		return errors.New("demo: login rejected by TCP server")
	}
	conn.SetDeadline(time.Time{})

	s.sessions[deviceID] = &demoSession{conn: conn, reader: reader}
	return nil
}

func (s *demoService) SendBurst(_ context.Context, deviceID string, count int) error {
	s.mu.RLock()
	sess, exists := s.sessions[deviceID]
	s.mu.RUnlock()
	if !exists {
		return ErrDemoSessionNotFound
	}

	for i := 0; i < count; i++ {
		sess.conn.SetDeadline(time.Now().Add(3 * time.Second))
		if _, err := sess.conn.Write([]byte(demo.RandomFrame())); err != nil {
			s.closeAndRemove(deviceID)
			return err
		}
		sess.reader.ReadString('#') //nolint:errcheck // ack is best-effort
	}
	sess.conn.SetDeadline(time.Time{})
	return nil
}

func (s *demoService) StopSession(_ context.Context, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, exists := s.sessions[deviceID]
	if !exists {
		return ErrDemoSessionNotFound
	}
	sess.conn.Close()
	delete(s.sessions, deviceID)
	return nil
}

func (s *demoService) IsActive(deviceID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.sessions[deviceID]
	return exists
}

func (s *demoService) closeAndRemove(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, exists := s.sessions[deviceID]; exists {
		sess.conn.Close()
		delete(s.sessions, deviceID)
	}
}
