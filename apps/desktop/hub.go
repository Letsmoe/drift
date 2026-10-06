package main

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Wire format, one JSON object per WebSocket text message:
//
//	request  {"id": 1, "method": "Search", "params": ["query"]}
//	response {"id": 1, "result": ...} or {"id": 1, "error": "message"}
//	event    {"event": "queue", "data": ...}
type rpcRequest struct {
	ID     int64             `json:"id"`
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

type rpcResponse struct {
	ID     int64  `json:"id"`
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

type rpcEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// hub serves /rpc and broadcasts events to every open connection.
type hub struct {
	dispatcher *dispatcher
	token      string
	logger     *slog.Logger
	upgrader   websocket.Upgrader

	mutex       sync.Mutex
	connections map[*rpcConnection]struct{}
}

// rpcConnection serializes writes; gorilla/websocket allows one writer at a time.
type rpcConnection struct {
	socket     *websocket.Conn
	writeMutex sync.Mutex
}

func newHub(token string, logger *slog.Logger) *hub {
	return &hub{
		token:  token,
		logger: logger,
		// Any origin may connect (the Vite dev server is a different origin); the
		// token is what keeps other local processes and web pages out.
		upgrader:    websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
		connections: make(map[*rpcConnection]struct{}),
	}
}

// serve sets the value whose methods requests call. Requests that arrive before
// it is set are rejected.
func (hub *hub) serve(target any) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	hub.dispatcher = newDispatcher(target)
}

func (hub *hub) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	given := request.URL.Query().Get("token")
	if subtle.ConstantTimeCompare([]byte(given), []byte(hub.token)) != 1 {
		http.Error(writer, "forbidden", http.StatusForbidden)
		return
	}
	socket, err := hub.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		hub.logger.Warn("rpc upgrade", "error", err)
		return
	}
	connection := &rpcConnection{socket: socket}
	hub.add(connection)
	defer hub.remove(connection)
	hub.readRequests(connection)
}

func (hub *hub) readRequests(connection *rpcConnection) {
	for {
		var request rpcRequest
		err := connection.socket.ReadJSON(&request)
		if err != nil {
			return
		}
		go hub.answer(connection, request)
	}
}

func (hub *hub) answer(connection *rpcConnection, request rpcRequest) {
	response := rpcResponse{ID: request.ID}
	result, err := hub.call(request)
	if err != nil {
		response.Error = err.Error()
	} else {
		response.Result = result
	}
	hub.write(connection, response)
}

func (hub *hub) call(request rpcRequest) (any, error) {
	hub.mutex.Lock()
	dispatcher := hub.dispatcher
	hub.mutex.Unlock()
	if dispatcher == nil {
		return nil, errBackendStarting
	}
	return dispatcher.call(request.Method, request.Params)
}

// broadcast sends an event to every connected frontend.
func (hub *hub) broadcast(name string, data any) {
	hub.mutex.Lock()
	connections := make([]*rpcConnection, 0, len(hub.connections))
	for connection := range hub.connections {
		connections = append(connections, connection)
	}
	hub.mutex.Unlock()

	for _, connection := range connections {
		hub.write(connection, rpcEvent{Event: name, Data: data})
	}
}

func (hub *hub) write(connection *rpcConnection, message any) {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()
	err := connection.socket.WriteJSON(message)
	if err != nil {
		hub.logger.Debug("rpc write", "error", err)
	}
}

func (hub *hub) add(connection *rpcConnection) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	hub.connections[connection] = struct{}{}
}

func (hub *hub) remove(connection *rpcConnection) {
	hub.mutex.Lock()
	delete(hub.connections, connection)
	hub.mutex.Unlock()
	_ = connection.socket.Close()
}
