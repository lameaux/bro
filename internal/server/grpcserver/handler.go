package grpcserver

import (
	"errors"
	"fmt"
	"io"

	"github.com/lameaux/bro/internal/server/restserver"
	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var emptyResponse = &pb.Empty{} //nolint:gochecknoglobals

func (s *server) Send(stream grpc.ClientStreamingServer[pb.Metric, pb.Empty]) error {
	for {
		metric, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil // client finished
		}

		if err != nil {
			return fmt.Errorf("failed to receive from client: %w", err)
		}

		log.Debug().
			Any("metric", metric).
			Msg("metric received")

		countFailedRequest("check")

		// Send response to client
		if err = stream.SendAndClose(emptyResponse); err != nil {
			return fmt.Errorf("failed to send to client: %w", err)
		}
	}
}

func requestLabels(scenarioName, method, url string) prometheus.Labels {
	return prometheus.Labels{
		"scenario": scenarioName,
		"method":   method,
		"url":      url,
	}
}

func countFailedRequest(reason string) {
	labels := requestLabels("scenario", "GET", "http://example.com")
	labels["reason"] = reason
	restserver.CountFailedRequest(labels)
}
