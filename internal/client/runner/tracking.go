package runner

import (
	"net/http"
	"strconv"
	"time"
)

func (r *Runner) trackError(err error) {
	for _, l := range r.listeners {
		l.TrackError(r.labels(nil), err)
	}
}

func (r *Runner) trackResponse(resp *http.Response, success bool, latency time.Duration) {
	for _, l := range r.listeners {
		l.TrackResponse(r.labels(resp), success, latency)
	}
}

func (r *Runner) labels(resp *http.Response) map[string]string {
	m := map[string]string{
		"scenario": r.scenario.Name,
		"method":   r.scenario.HttpRequest.Method(),
		"url":      r.scenario.HttpRequest.Url,
	}

	if resp != nil {
		m["code"] = strconv.Itoa(resp.StatusCode)
	}

	return m
}
