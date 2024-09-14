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

type BrodWorker struct {
	Conn       *grpc.ClientConn
	InstanceID string
}

func NewWorker(addr string) (*BrodWorker, error) {
	conn, err := grpc_client.GrpcConnection(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v: %w", addr, err)
	}

	w := &BrodWorker{
		Conn:       conn,
		InstanceID: uuid.NewString(),
	}

	return w, nil
}

func (b *BrodWorker) Run(ctx context.Context) {
	defer b.Conn.Close()

	log.Debug().
		Str("instance", b.InstanceID).
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

func (b *BrodWorker) sendCounters(ctx context.Context) error {
	c := pb.NewMetricsClient(b.Conn)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	r, err := c.SendCounters(ctxWithTimeout, &pb.Counters{
		Id: b.InstanceID,
	})

	if err != nil {
		return fmt.Errorf("failed to send counters: %w", err)
	}

	log.Debug().Str("msg", r.Msg).Msg("counters sent")

	return nil
}
