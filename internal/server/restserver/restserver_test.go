package restserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lameaux/bro/internal/server/restserver"
	"github.com/lameaux/bro/internal/shared/banner"
)

func TestIndexHandler(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	recorder := httptest.NewRecorder()

	handler := restserver.IndexHandler()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %v", recorder.Code)
	}

	if recorder.Body.String() != banner.Banner {
		t.Errorf("expected banner, got %q", recorder.Body.String())
	}
}
