package stats

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lameaux/bro/internal/client/grpcclient"
	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Sender struct {
	conn *grpc.ClientConn

	instanceID string
	groupID    string

	queue []TrackInfo
	mu    sync.RWMutex
}

type TrackInfo struct {
	Labels  map[string]string
	Error   bool
	Timeout bool
	Success bool
	Latency time.Duration
}

func NewSender(serverAddr, groupID string) (*Sender, error) {
	conn, err := grpcclient.GrpcConnection(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v: %w", serverAddr, err)
	}

	worker := &Sender{
		conn:       conn,
		instanceID: uuid.NewString(),
		groupID:    groupID,
	}

	return worker, nil
}

func (s *Sender) Run(ctx context.Context) {
	defer s.conn.Close()

	log.Debug().
		Str("instance", s.instanceID).
		Msg("started stats worker")

	rateTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-rateTicker.C:
			if err := s.sendTracking(ctx); err != nil {
				log.Warn().Err(err).Msg("failed to send tracking")
			}
		}
	}
}

func (s *Sender) sendTracking(ctx context.Context) error {
	c := pb.NewMetricsClient(s.conn)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	stream, err := c.Send(ctxWithTimeout)
	if err != nil {
		return fmt.Errorf("failed to make grpc call: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, info := range s.queue {
		err = stream.Send(&pb.Metric{
			Id:      s.instanceID,
			Group:   s.groupID,
			Labels:  info.Labels,
			Success: info.Success,
			Error:   info.Error,
			Timeout: info.Timeout,
			Latency: info.Latency.Milliseconds(),
		})
		if err != nil {
			return fmt.Errorf("failed to send metric: %w", err)
		}
	}

	s.queue = nil

	if err = stream.CloseSend(); err != nil {
		return fmt.Errorf("failed to close sending: %w", err)
	}

	log.Debug().Msg("tracking sent")

	return nil
}

func (s *Sender) TrackError(labels map[string]string, err error) {
	var netErr net.Error
	timeout := errors.As(err, &netErr) && netErr.Timeout()

	info := TrackInfo{
		Labels:  labels,
		Error:   true,
		Timeout: timeout,
		Success: false,
		Latency: 0,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.queue = append(s.queue, info)
}

func (s *Sender) TrackResponse(labels map[string]string, success bool, latency time.Duration) {
	info := TrackInfo{
		Labels:  labels,
		Error:   false,
		Timeout: false,
		Success: success,
		Latency: latency,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.queue = append(s.queue, info)
}
