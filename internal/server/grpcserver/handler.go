package grpcserver

import (
	"context"

	"github.com/lameaux/bro/internal/server/restserver"
	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

func (s *server) SendCounters(_ context.Context, counters *pb.Counters) (*pb.Result, error) {
	log.Debug().
		Str("instance", counters.GetId()).
		Msg("counter received")

	countFailedRequest("check")

	//nolint:godox
	// FIXME: I know about it
	// labels := r.responseLabels(response, success)
	// metrics.HttpResponsesTotal.With(labels).Inc()
	//
	// labels = r.responseLabels(response, success)
	// metrics.HttpRequestDurationSec.With(labels).Observe(latency.Seconds())

	return &pb.Result{Msg: "received id=" + counters.GetId()}, nil
}

func requestLabels(scenarioName, method, url string) prometheus.Labels {
	return prometheus.Labels{
		"scenario": scenarioName,
		"method":   method,
		"url":      url,
	}
}

// func responseLabels(response *http.Response, success bool) prometheus.Labels {
//	labels := requestLabels("scenario", "GET", "http://example.com")
//	labels["code"] = strconv.Itoa(response.StatusCode)
//	labels["success"] = strconv.FormatBool(success)
//
//	return labels
//}

func countFailedRequest(reason string) {
	labels := requestLabels("scenario", "GET", "http://example.com")
	labels["reason"] = reason
	restserver.CountFailedRequest(labels)
}
