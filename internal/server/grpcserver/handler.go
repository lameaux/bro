package grpcserver

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var emptyResponse = &pb.Empty{} //nolint:gochecknoglobals

func (s *server) Send(stream grpc.ClientStreamingServer[pb.MetricV1, pb.Empty]) error {
	for {
		metric, err := stream.Recv()
		if errors.Is(err, io.EOF) { // client finished
			// Send response to client
			if err = stream.SendAndClose(emptyResponse); err != nil {
				return fmt.Errorf("failed to send to client: %w", err)
			}

			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to receive from client: %w", err)
		}

		log.Debug().
			Any("metric", metric).
			Msg("metric received")

		s.countRequestMetric(metric)
	}
}

func (s *server) countRequestMetric(metric *pb.MetricV1) {
	labels := requestLabels(metric)
	s.promMetrics.CountRequest(labels, metric.GetLatencySeconds())
}

func requestLabels(metric *pb.MetricV1) map[string]string {
	return map[string]string{
		"instance_id": metric.GetInstance(),
		"group_id":    metric.GetGroup(),
		"scenario":    metric.GetScenario(),
		"method":      metric.GetMethod(),
		"url":         metric.GetUrl(),
		"code":        metric.GetCode(),
		"failed":      strconv.FormatBool(metric.GetFailed()),
		"timeout":     strconv.FormatBool(metric.GetTimeout()),
		"success":     strconv.FormatBool(metric.GetSuccess()),
	}
}
