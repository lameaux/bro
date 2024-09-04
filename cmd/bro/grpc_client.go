package main

import (
	"context"
	"fmt"
	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

func grpcConnection(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func sendCounters(ctx context.Context, addr string) error {
	conn, err := grpcConnection(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %v: %w", addr, err)
	}
	defer conn.Close()

	c := pb.NewMetricsClient(conn)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	r, err := c.SendCounters(ctxWithTimeout, &pb.Counters{
		Name:  "now",
		Value: time.Now().String(),
	})

	if err != nil {
		return fmt.Errorf("failed to send counters: %w", err)
	}

	log.Debug().Str("msg", r.Msg).Msg("counters sent")

	return nil
}
