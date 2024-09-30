package runner

import (
	"net/http"
	"strconv"
	"time"

	"github.com/lameaux/bro/internal/client/tracking"
)

func (r *Runner) trackError(err error) {
	for _, l := range r.listeners {
		l.TrackFailed(r.requestInfo(nil), err)
	}
}

func (r *Runner) trackResponse(resp *http.Response, success bool, latency time.Duration) {
	for _, l := range r.listeners {
		l.TrackResponse(r.requestInfo(resp), success, latency)
	}
}

func (r *Runner) requestInfo(resp *http.Response) *tracking.RequestInfo {
	info := &tracking.RequestInfo{
		Scenario: r.scenario.Name,
		Method:   r.scenario.HTTPRequest.Method(),
		URL:      r.scenario.HTTPRequest.URL,
	}

	if resp != nil {
		info.Code = strconv.Itoa(resp.StatusCode)
	}

	return info
}
