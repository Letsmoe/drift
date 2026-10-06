package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

type fakeService struct{}

func (fakeService) Nothing()      {}
func (fakeService) Value() string { return "value" }
func (fakeService) Fails() error  { return errors.New("failed") }
func (fakeService) Sum(left, right int) (int, error) {
	if left < 0 {
		return 0, errors.New("negative")
	}
	return left + right, nil
}
func (fakeService) Panics() string { panic("boom") }

func rawParams(t *testing.T, values ...any) []json.RawMessage {
	t.Helper()
	params := make([]json.RawMessage, len(values))
	for index, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		params[index] = encoded
	}
	return params
}

func TestDispatcherReturnShapes(t *testing.T) {
	dispatcher := newDispatcher(fakeService{})
	tests := []struct {
		method    string
		params    []json.RawMessage
		want      any
		wantError string
	}{
		{method: "Nothing", want: nil},
		{method: "Value", want: "value"},
		{method: "Fails", wantError: "failed"},
		{method: "Sum", params: rawParams(t, 2, 3), want: 5},
		{method: "Sum", params: rawParams(t, 2), wantError: "takes 2 arguments, got 1"},
		{method: "Sum", params: rawParams(t, "2", 3), wantError: "argument 1"},
		{method: "Panics", wantError: "Panics panicked: boom"},
		{method: "Missing", wantError: `unknown method "Missing"`},
	}
	for _, test := range tests {
		got, err := dispatcher.call(test.method, test.params)
		if test.wantError != "" {
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("%s: error %v, want it to contain %q", test.method, err, test.wantError)
			}
			continue
		}
		if err != nil || got != test.want {
			t.Errorf("%s: got %v, %v; want %v", test.method, got, err, test.want)
		}
	}
}

func startHub(t *testing.T) (*hub, string) {
	t.Helper()
	rpc := newHub("secret", slog.New(slog.NewTextHandler(io.Discard, nil)))
	rpc.serve(fakeService{})
	server := httptest.NewServer(rpc)
	t.Cleanup(server.Close)
	return rpc, "ws" + strings.TrimPrefix(server.URL, "http")
}

func closeBody(t *testing.T, response *http.Response) {
	t.Helper()
	err := response.Body.Close()
	if err != nil {
		t.Error(err)
	}
}

func TestHubRejectsWrongToken(t *testing.T) {
	_, url := startHub(t)
	_, response, err := websocket.DefaultDialer.Dial(url+"?token=wrong", nil)
	if err == nil {
		t.Fatal("connected with a wrong token")
	}
	defer closeBody(t, response)
	if response.StatusCode != http.StatusForbidden {
		t.Errorf("status %d, want %d", response.StatusCode, http.StatusForbidden)
	}
}

func TestHubAnswersCallsAndBroadcastsEvents(t *testing.T) {
	rpc, url := startHub(t)
	socket, response, err := websocket.DefaultDialer.Dial(url+"?token=secret", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer closeBody(t, response)
	defer func() { _ = socket.Close() }()

	err = socket.WriteJSON(rpcRequest{ID: 7, Method: "Sum", Params: rawParams(t, 2, 3)})
	if err != nil {
		t.Fatal(err)
	}
	var answer struct {
		ID     int64 `json:"id"`
		Result int   `json:"result"`
	}
	err = socket.ReadJSON(&answer)
	if err != nil || answer.ID != 7 || answer.Result != 5 {
		t.Fatalf("answer %+v, %v; want id 7, result 5", answer, err)
	}

	rpc.broadcast("queue", "changed")
	var event rpcEvent
	err = socket.ReadJSON(&event)
	if err != nil || event.Event != "queue" || event.Data != "changed" {
		t.Fatalf("event %+v, %v; want queue/changed", event, err)
	}
}
