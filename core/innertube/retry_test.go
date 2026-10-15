package innertube

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// refusingServer answers the first refusals requests with status, then a visitor ID.
func refusingServer(t *testing.T, status int, refusals int32) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	requests := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) <= refusals {
			writer.WriteHeader(status)
			return
		}
		_, _ = writer.Write([]byte(`{"responseContext":{"visitorData":"` + fakeVisitorData + `"}}`))
	}))
	t.Cleanup(server.Close)
	return server, requests
}

func clientWithoutRetryPause(baseURL string) *Client {
	client := NewClient(WithBaseURL(baseURL))
	client.retryDelays = []time.Duration{0, 0}
	return client
}

func TestRetriesATransientRefusal(t *testing.T) {
	server, requests := refusingServer(t, http.StatusForbidden, 1)
	visitorData, err := clientWithoutRetryPause(server.URL).VisitorData(t.Context())
	if err != nil {
		t.Fatalf("visitor data after one 403: %v", err)
	}
	if visitorData != fakeVisitorData || requests.Load() != 2 {
		t.Errorf("visitor data %q after %d requests, want the fake one after 2", visitorData, requests.Load())
	}
}

func TestGivesUpAfterTheLastRetry(t *testing.T) {
	server, requests := refusingServer(t, http.StatusForbidden, 10)
	_, err := clientWithoutRetryPause(server.URL).VisitorData(t.Context())
	if !isTransient(err) {
		t.Fatalf("err = %v, want the 403", err)
	}
	if requests.Load() != 3 {
		t.Errorf("requests = %d, want 3 (one try, two retries)", requests.Load())
	}
}

func TestDoesNotRetryAPermanentStatus(t *testing.T) {
	server, requests := refusingServer(t, http.StatusNotFound, 10)
	_, err := clientWithoutRetryPause(server.URL).VisitorData(t.Context())
	if err == nil || requests.Load() != 1 {
		t.Errorf("err = %v after %d requests, want a 404 after 1", err, requests.Load())
	}
}
