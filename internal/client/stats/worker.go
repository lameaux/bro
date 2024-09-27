package stats

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lameaux/bro/internal/client/grpcclient"
	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Worker struct {
	conn *grpc.ClientConn

	instanceID string
	groupID    string

	counters *RequestCounters
}

func NewWorker(serverAddr, groupID string) (*Worker, error) {
	conn, err := grpcclient.GrpcConnection(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v: %w", serverAddr, err)
	}

	counters := NewRequestCounters()

	worker := &Worker{
		conn:       conn,
		instanceID: uuid.NewString(),
		groupID:    groupID,
		counters:   counters,
	}

	return worker, nil
}

func (b *Worker) Run(ctx context.Context) {
	defer b.conn.Close()

	log.Debug().
		Str("instance", b.instanceID).
		Msg("started brod worker")

	rateTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-rateTicker.C:
			if err := b.sendCounters(ctx); err != nil {
				log.Warn().Err(err).Msg("failed to send counters")
			}
		}
	}
}

func (b *Worker) sendCounters(ctx context.Context) error {
	c := pb.NewMetricsClient(b.conn)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	result, err := c.SendCounters(ctxWithTimeout, &pb.Counters{
		Id: b.instanceID,
	})
	if err != nil {
		return fmt.Errorf("failed to send counters: %w", err)
	}

	log.Debug().Str("msg", result.GetMsg()).Msg("counters sent")

	return nil
}

func (b *Worker) Counters() *RequestCounters {
	return b.counters
}
