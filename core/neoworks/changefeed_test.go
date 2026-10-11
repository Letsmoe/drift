package neoworks

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeFeedSession is a signed-in (or out) user pointing at a test server.
type fakeFeedSession struct {
	signedIn atomic.Bool
	url      string
	client   *http.Client
}

func newFakeFeedSession(url string) *fakeFeedSession {
	session := &fakeFeedSession{url: url, client: &http.Client{Timeout: time.Minute}}
	session.signedIn.Store(true)
	return session
}

func (session *fakeFeedSession) SignedIn() bool { return session.signedIn.Load() }

func (session *fakeFeedSession) accessToken(context.Context) (string, error) {
	if !session.SignedIn() {
		return "", ErrSignedOut
	}
	return "access-token", nil
}

func (session *fakeFeedSession) changesURL() string { return session.url }

func (session *fakeFeedSession) streamClient() *http.Client { return session.client }

// recordingChanges logs what the feed told the syncer.
type recordingChanges struct {
	mutex sync.Mutex
	notes []string
}

func (changes *recordingChanges) record(note string) {
	changes.mutex.Lock()
	defer changes.mutex.Unlock()
	changes.notes = append(changes.notes, note)
}

func (changes *recordingChanges) NoteRemoteChange(sequence int64) {
	changes.record(fmt.Sprintf("change %d", sequence))
}
func (changes *recordingChanges) NoteResync() { changes.record("resync") }
func (changes *recordingChanges) Request()    { changes.record("request") }

func (changes *recordingChanges) snapshot() []string {
	changes.mutex.Lock()
	defer changes.mutex.Unlock()
	return append([]string(nil), changes.notes...)
}

func (changes *recordingChanges) count(note string) int {
	total := 0
	for _, recorded := range changes.snapshot() {
		if recorded == note {
			total++
		}
	}
	return total
}

// feedServer answers each connection with the next script entry and records
// the requests. The last entry answers every later connection.
type feedServer struct {
	server   *httptest.Server
	mutex    sync.Mutex
	requests []*http.Request
	times    []time.Time
	script   []func(writer http.ResponseWriter, request *http.Request)
}

func newFeedServer(t *testing.T, script ...func(http.ResponseWriter, *http.Request)) *feedServer {
	t.Helper()
	feed := &feedServer{script: script}
	feed.server = httptest.NewServer(http.HandlerFunc(feed.serve))
	t.Cleanup(feed.server.Close)
	return feed
}

func (feed *feedServer) serve(writer http.ResponseWriter, request *http.Request) {
	feed.mutex.Lock()
	index := min(len(feed.requests), len(feed.script)-1)
	feed.requests = append(feed.requests, request)
	feed.times = append(feed.times, time.Now())
	handler := feed.script[index]
	feed.mutex.Unlock()
	handler(writer, request)
}

func (feed *feedServer) connections() int {
	feed.mutex.Lock()
	defer feed.mutex.Unlock()
	return len(feed.requests)
}

func (feed *feedServer) request(index int) *http.Request {
	feed.mutex.Lock()
	defer feed.mutex.Unlock()
	return feed.requests[index]
}

// sendThenEnd writes the text as an event stream and ends the response.
func sendThenEnd(text string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = writer.Write([]byte(text))
	}
}

// sendThenHold writes the text and keeps the stream open until the client leaves.
func sendThenHold(text string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = writer.Write([]byte(text))
		writer.(http.Flusher).Flush()
		<-request.Context().Done()
	}
}

func refuse(status int) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(status)
	}
}

func startFeed(t *testing.T, session *fakeFeedSession, changes *recordingChanges, configure func(*ChangeFeedConfig)) *ChangeFeed {
	t.Helper()
	config := ChangeFeedConfig{
		Session: session, Changes: changes,
		Backoff: 5 * time.Millisecond, MaxBackoff: 20 * time.Millisecond, IdleTimeout: time.Minute,
	}
	if configure != nil {
		configure(&config)
	}
	feed := NewChangeFeed(config)
	feed.Start(t.Context())
	t.Cleanup(feed.Stop)
	return feed
}

func waitUntil(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}

func TestChangeEventsReachTheSyncerWithTheSessionsToken(t *testing.T) {
	server := newFeedServer(t, sendThenHold("retry: 3000\n\n: connected\n\nid: 7\nevent: change\ndata: {\"collection\":\"@neoworks/drift\",\"seq\":7}\n\n"))
	changes := &recordingChanges{}
	startFeed(t, newFakeFeedSession(server.server.URL), changes, nil)

	waitUntil(t, "the change", func() bool { return changes.count("change 7") == 1 })

	request := server.request(0)
	if got := request.Header.Get("Authorization"); got != "Bearer access-token" {
		t.Errorf("Authorization = %q", got)
	}
	if got := request.Header.Get("Accept"); got != "text/event-stream" {
		t.Errorf("Accept = %q", got)
	}
	if got := request.Header.Get("Last-Event-ID"); got != "" {
		t.Errorf("a first connection has nothing to resume, sent Last-Event-ID %q", got)
	}
}

func TestResyncAndUnreadableEventsAskForAPull(t *testing.T) {
	server := newFeedServer(t, sendThenHold(
		"event: resync\ndata: {\"reason\":\"lost\"}\n\n"+
			"id: 9\nevent: change\ndata: not json\n\n"+
			"event: change\ndata: {\"collection\":\"c\",\"seq\":0}\n\n"))
	changes := &recordingChanges{}
	startFeed(t, newFakeFeedSession(server.server.URL), changes, nil)

	waitUntil(t, "three resyncs", func() bool { return changes.count("resync") == 3 })

	if changes.count("change 9") != 0 {
		t.Errorf("an unreadable change must not count as a sequence: %v", changes.snapshot())
	}
}

func TestEveryConnectAsksForASync(t *testing.T) {
	server := newFeedServer(t, sendThenEnd(": hello\n\n"), sendThenHold(": hello\n\n"))
	changes := &recordingChanges{}
	startFeed(t, newFakeFeedSession(server.server.URL), changes, nil)

	waitUntil(t, "two connections", func() bool { return server.connections() == 2 })

	waitUntil(t, "a sync request for each", func() bool { return changes.count("request") == 2 })
}

func TestReconnectingResumesFromTheLastEventID(t *testing.T) {
	server := newFeedServer(t,
		sendThenEnd("id: 3\nevent: change\ndata: {\"collection\":\"c\",\"seq\":3}\n\n"),
		sendThenEnd("id: 8\nevent: change\ndata: {\"collection\":\"c\",\"seq\":8}\n\n"),
		sendThenHold(": hello\n\n"))
	changes := &recordingChanges{}
	startFeed(t, newFakeFeedSession(server.server.URL), changes, nil)

	waitUntil(t, "three connections", func() bool { return server.connections() == 3 })

	if got := server.request(1).Header.Get("Last-Event-ID"); got != "3" {
		t.Errorf("second connection Last-Event-ID = %q, want 3", got)
	}
	if got := server.request(2).Header.Get("Last-Event-ID"); got != "8" {
		t.Errorf("third connection Last-Event-ID = %q, want 8", got)
	}
}

func TestFailedConnectionsBackOffAndRecover(t *testing.T) {
	server := newFeedServer(t, refuse(http.StatusBadGateway), refuse(http.StatusServiceUnavailable), refuse(http.StatusNotFound),
		sendThenHold("id: 4\nevent: change\ndata: {\"collection\":\"c\",\"seq\":4}\n\n"))
	changes := &recordingChanges{}
	startFeed(t, newFakeFeedSession(server.server.URL), changes, nil)

	waitUntil(t, "the change after three refusals", func() bool { return changes.count("change 4") == 1 })

	if got := changes.count("request"); got != 1 {
		t.Errorf("only a connection that opened asks for a sync, got %d requests", got)
	}
}

func TestRetryAfterSetsTheWait(t *testing.T) {
	limited := func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Retry-After", "1")
		writer.WriteHeader(http.StatusTooManyRequests)
	}
	server := newFeedServer(t, limited, sendThenHold(": hello\n\n"))
	startFeed(t, newFakeFeedSession(server.server.URL), &recordingChanges{}, nil)

	waitUntil(t, "the second connection", func() bool { return server.connections() == 2 })

	if gap := server.times[1].Sub(server.times[0]); gap < 900*time.Millisecond {
		t.Errorf("reconnected after %v, the server asked for a second", gap)
	}
}

func TestAStreamThatGoesQuietIsDroppedAndReopened(t *testing.T) {
	server := newFeedServer(t, sendThenHold(": hello\n\n"))
	startFeed(t, newFakeFeedSession(server.server.URL), &recordingChanges{}, func(config *ChangeFeedConfig) {
		config.IdleTimeout = 60 * time.Millisecond
	})

	waitUntil(t, "a reconnect after the silence", func() bool { return server.connections() >= 2 })
}

func TestHeartbeatsKeepAStreamOpen(t *testing.T) {
	heartbeats := func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		for range 12 {
			_, _ = writer.Write([]byte(": heartbeat\n\n"))
			writer.(http.Flusher).Flush()
			select {
			case <-time.After(20 * time.Millisecond):
			case <-request.Context().Done():
				return
			}
		}
		<-request.Context().Done()
	}
	server := newFeedServer(t, heartbeats)
	startFeed(t, newFakeFeedSession(server.server.URL), &recordingChanges{}, func(config *ChangeFeedConfig) {
		config.IdleTimeout = 100 * time.Millisecond
	})

	time.Sleep(300 * time.Millisecond)

	if got := server.connections(); got != 1 {
		t.Errorf("heartbeats should keep one connection open, saw %d", got)
	}
}

func TestNothingConnectsWhileSignedOutAndAWakeConnectsAfterSignIn(t *testing.T) {
	server := newFeedServer(t, sendThenHold(": hello\n\n"))
	session := newFakeFeedSession(server.server.URL)
	session.signedIn.Store(false)
	feed := startFeed(t, session, &recordingChanges{}, nil)

	time.Sleep(50 * time.Millisecond)
	if got := server.connections(); got != 0 {
		t.Fatalf("a signed-out feed connected %d times", got)
	}

	session.signedIn.Store(true)
	feed.Wake()

	waitUntil(t, "the connection after sign-in", func() bool { return server.connections() == 1 })
}

func TestSigningOutClosesTheStreamAndForgetsTheEventID(t *testing.T) {
	var closedStreams atomic.Int32
	hold := func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = writer.Write([]byte("id: 5\nevent: change\ndata: {\"collection\":\"c\",\"seq\":5}\n\n"))
		writer.(http.Flusher).Flush()
		<-request.Context().Done()
		closedStreams.Add(1)
	}
	server := newFeedServer(t, hold)
	session := newFakeFeedSession(server.server.URL)
	changes := &recordingChanges{}
	feed := startFeed(t, session, changes, nil)
	waitUntil(t, "the first change", func() bool { return changes.count("change 5") == 1 })

	session.signedIn.Store(false)
	feed.Wake()
	waitUntil(t, "the server to see the stream close", func() bool { return closedStreams.Load() == 1 })
	time.Sleep(50 * time.Millisecond)
	if got := server.connections(); got != 1 {
		t.Fatalf("a signed-out feed reconnected: %d connections", got)
	}

	session.signedIn.Store(true)
	feed.Wake()
	waitUntil(t, "the connection after signing in again", func() bool { return server.connections() == 2 })
	if got := server.request(1).Header.Get("Last-Event-ID"); got != "" {
		t.Errorf("another sign-in must not resume the old account's stream, sent %q", got)
	}
}

func TestStopClosesTheStream(t *testing.T) {
	var closed atomic.Bool
	hold := func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = writer.Write([]byte(": hello\n\n"))
		writer.(http.Flusher).Flush()
		<-request.Context().Done()
		closed.Store(true)
	}
	server := newFeedServer(t, hold)
	changes := &recordingChanges{}
	feed := NewChangeFeed(ChangeFeedConfig{Session: newFakeFeedSession(server.server.URL), Changes: changes})
	feed.Start(t.Context())
	waitUntil(t, "the connection", func() bool { return server.connections() == 1 })

	stopped := make(chan struct{})
	go func() {
		feed.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("Stop did not return")
	}
	waitUntil(t, "the server to see the stream close", closed.Load)
}

func TestSessionStreamsWithoutARequestTimeout(t *testing.T) {
	session := &Session{config: Config{APIURL: "https://api.neoworks.dev", HTTPClient: &http.Client{Timeout: time.Minute}}}

	if got := session.changesURL(); got != "https://api.neoworks.dev/api/v1/changes" {
		t.Errorf("changesURL = %q", got)
	}
	if got := session.streamClient().Timeout; got != 0 {
		t.Errorf("a stream cannot have a timeout, got %v", got)
	}
	if session.config.HTTPClient.Timeout != time.Minute {
		t.Error("the session's own client must keep its timeout")
	}
}

func TestEventReaderReadsFieldsCommentsAndMultilineData(t *testing.T) {
	text := ": connected\n\nretry: 3000\n\nid: 12\nevent: change\ndata: {\"a\":\ndata: 1}\n\nevent: empty\n\ndata: bare\n\n"
	lines := 0
	reader := newEventReader(strings.NewReader(text), func() { lines++ })

	first, err := reader.next()
	if err != nil {
		t.Fatal(err)
	}
	if first != (serverEvent{id: "12", hasID: true, name: "change", data: "{\"a\":\n1}"}) {
		t.Errorf("first = %+v", first)
	}
	second, err := reader.next()
	if err != nil {
		t.Fatal(err)
	}
	if second != (serverEvent{data: "bare"}) {
		t.Errorf("an event without data is dropped and ids do not leak, got %+v", second)
	}
	if _, err = reader.next(); err == nil {
		t.Error("the end of the stream is an error")
	}
	if lines == 0 {
		t.Error("onLine was never called")
	}
}
