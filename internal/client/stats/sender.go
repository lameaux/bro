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
	"github.com/lameaux/bro/internal/client/tracking"
	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Sender struct {
	conn *grpc.ClientConn

	instance string
	group    string

	queue []TrackInfo
	mu    sync.Mutex
}

type TrackInfo struct {
	Scenario string
	Method   string
	URL      string
	Code     string

	Failed  bool
	Timeout bool
	Success bool

	Latency time.Duration
}

func NewSender(serverAddr, group string) (*Sender, error) {
	conn, err := grpcclient.GrpcConnection(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v: %w", serverAddr, err)
	}

	worker := &Sender{
		conn:     conn,
		instance: uuid.NewString(),
		group:    group,
	}

	return worker, nil
}

func (s *Sender) Run(ctx context.Context) {
	defer s.conn.Close()

	log.Debug().
		Str("instance", s.instance).
		Str("group", s.group).
		Msg("started stats sender")

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
	c := pb.NewMetricsV1Client(s.conn)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	stream, err := c.Send(ctxWithTimeout)
	if err != nil {
		return fmt.Errorf("failed to make grpc call: %w", err)
	}

	s.mu.Lock()
	queueCopy := s.queue
	s.queue = nil
	s.mu.Unlock()

	for _, info := range queueCopy {
		err = stream.Send(&pb.MetricV1{
			Instance: s.instance,
			Group:    s.group,

			Scenario: info.Scenario,
			Method:   info.Method,
			Url:      info.URL,
			Code:     info.Code,

			Failed:  info.Failed,
			Timeout: info.Timeout,
			Success: info.Success,

			LatencySeconds: info.Latency.Seconds(),
		})
		if err != nil {
			return fmt.Errorf("failed to send metric: %w", err)
		}
	}

	if _, err = stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("failed to close stream: %w", err)
	}

	log.Debug().Int("count", len(queueCopy)).Msg("tracking sent")

	return nil
}

func (s *Sender) TrackFailed(
	reqInfo *tracking.RequestInfo,
	err error,
) {
	var netErr net.Error
	timeout := errors.As(err, &netErr) && netErr.Timeout()

	info := TrackInfo{
		Scenario: reqInfo.Scenario,
		Method:   reqInfo.Method,
		URL:      reqInfo.URL,
		Failed:   true,
		Timeout:  timeout,
	}

	s.mu.Lock()
	s.queue = append(s.queue, info)
	s.mu.Unlock()
}

func (s *Sender) TrackResponse(
	reqInfo *tracking.RequestInfo,
	success bool,
	latency time.Duration,
) {
	info := TrackInfo{
		Scenario: reqInfo.Scenario,
		Method:   reqInfo.Method,
		URL:      reqInfo.URL,
		Code:     reqInfo.Code,
		Success:  success,
		Latency:  latency,
	}

	s.mu.Lock()
	s.queue = append(s.queue, info)
	s.mu.Unlock()
}
