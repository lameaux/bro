package grpc_server

import (
	"context"
	"fmt"
	"github.com/lameaux/bro/internal/server/metrics"
	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

func (s *server) SendCounters(_ context.Context, counters *pb.Counters) (*pb.Result, error) {
	log.Debug().
		Str("instance", counters.Id).
		Msg("counter received")

	countFailedRequest("check")

	//labels := r.responseLabels(response, success)
	//metrics.HttpResponsesTotal.With(labels).Inc()
	//
	//labels = r.responseLabels(response, success)
	//metrics.HttpRequestDurationSec.With(labels).Observe(latency.Seconds())

	return &pb.Result{Msg: fmt.Sprintf("received id=%s", counters.Id)}, nil
}

func requestLabels(scenarioName, method, url string) prometheus.Labels {
	return prometheus.Labels{
		"scenario": scenarioName,
		"method":   method,
		"url":      url,
	}
}

func responseLabels(response *http.Response, success bool) prometheus.Labels {
	labels := requestLabels("scenario", "GET", "http://example.com")
	labels["code"] = strconv.Itoa(response.StatusCode)
	labels["success"] = strconv.FormatBool(success)

	return labels
}

func countFailedRequest(reason string) {
	labels := requestLabels("scenario", "GET", "http://example.com")
	labels["reason"] = reason
	metrics.HttpRequestsFailedTotal.With(labels).Inc()
}
