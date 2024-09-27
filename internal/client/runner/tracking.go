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
	labelsMap := map[string]string{
		"scenario": r.scenario.Name,
		"method":   r.scenario.HTTPRequest.Method(),
		"url":      r.scenario.HTTPRequest.URL,
	}

	if resp != nil {
		labelsMap["code"] = strconv.Itoa(resp.StatusCode)
	}

	return labelsMap
}
