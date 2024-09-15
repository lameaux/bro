package stats

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lameaux/bro/internal/client/grpc_client"
	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"time"
)

type Worker struct {
	conn *grpc.ClientConn

	instanceID string
	groupID    string

	counters *RequestCounters
}

func NewWorker(serverAddr, groupId string) (*Worker, error) {
	conn, err := grpc_client.GrpcConnection(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v: %w", serverAddr, err)
	}

	counters := NewRequestCounters()

	w := &Worker{
		conn:       conn,
		instanceID: uuid.NewString(),
		groupID:    groupId,
		counters:   counters,
	}

	return w, nil
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

	r, err := c.SendCounters(ctxWithTimeout, &pb.Counters{
		Id: b.instanceID,
	})

	if err != nil {
		return fmt.Errorf("failed to send counters: %w", err)
	}

	log.Debug().Str("msg", r.Msg).Msg("counters sent")

	return nil
}

func (b *Worker) Counters() *RequestCounters {
	return b.counters
}
