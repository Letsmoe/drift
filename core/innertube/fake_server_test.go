package innertube

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

const fakeVisitorData = "visitor-from-fake-server"

// fakeServer answers InnerTube endpoints with fixtures and records what it received.
type fakeServer struct {
	*httptest.Server

	mutex     sync.Mutex
	responses map[string][]byte
	requests  map[string][]recordedRequest
}

type recordedRequest struct {
	headers http.Header
	body    map[string]any
}

func newFakeServer(t *testing.T) *fakeServer {
	t.Helper()
	server := &fakeServer{
		responses: map[string][]byte{
			"visitor_id": []byte(`{"responseContext":{"visitorData":"` + fakeVisitorData + `"}}`),
		},
		requests: map[string][]recordedRequest{},
	}
	server.Server = httptest.NewServer(http.HandlerFunc(server.handle))
	t.Cleanup(server.Close)
	return server
}

func (server *fakeServer) respondWithFixture(t *testing.T, endpoint, fixturePath string) {
	t.Helper()
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.responses[endpoint] = fixture
}

func (server *fakeServer) handle(writer http.ResponseWriter, request *http.Request) {
	endpoint := strings.TrimPrefix(request.URL.Path, "/youtubei/v1/")

	var body map[string]any
	_ = json.NewDecoder(request.Body).Decode(&body)

	server.mutex.Lock()
	server.requests[endpoint] = append(server.requests[endpoint], recordedRequest{headers: request.Header, body: body})
	response, found := server.responses[endpoint]
	server.mutex.Unlock()

	if !found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	_, _ = writer.Write(response)
}

func (server *fakeServer) lastRequest(t *testing.T, endpoint string) recordedRequest {
	t.Helper()
	server.mutex.Lock()
	defer server.mutex.Unlock()
	requests := server.requests[endpoint]
	if len(requests) == 0 {
		t.Fatalf("no request to %s", endpoint)
	}
	return requests[len(requests)-1]
}

func (server *fakeServer) requestCount(endpoint string) int {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return len(server.requests[endpoint])
}

func (server *fakeServer) client() *Client {
	return NewClient(WithBaseURL(server.URL))
}

// contextClient returns body.context.client of a recorded request.
func (request recordedRequest) contextClient() map[string]any {
	requestContext, _ := request.body["context"].(map[string]any)
	client, _ := requestContext["client"].(map[string]any)
	return client
}
