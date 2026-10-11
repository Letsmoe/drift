package neoworks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	changesPath = "/api/v1/changes"
	// DefaultFeedBackoff is the wait before reconnecting after a stream ended or failed once.
	DefaultFeedBackoff = time.Second
	// DefaultFeedMaxBackoff caps the wait between failed attempts.
	DefaultFeedMaxBackoff = 5 * time.Minute
	// DefaultFeedIdleTimeout is how long a stream may send nothing, heartbeats
	// included, before it is dropped; the API sends one every 20 seconds.
	DefaultFeedIdleTimeout = time.Minute
	// signedOutPoll is how often a feed with nobody signed in looks again.
	signedOutPoll = 10 * time.Second
)

// remoteChanges is what the feed tells the Syncer.
type remoteChanges interface {
	// NoteRemoteChange reports a change with the feed's sequence.
	NoteRemoteChange(sequence int64)
	// NoteResync reports that the feed cannot say what changed.
	NoteResync()
	// Request schedules a pass.
	Request()
}

// feedSession is the part of Session a ChangeFeed uses.
type feedSession interface {
	SignedIn() bool
	accessToken(ctx context.Context) (string, error)
	changesURL() string
	streamClient() *http.Client
}

// ChangeFeedConfig holds what a ChangeFeed needs; the durations default to the package defaults.
type ChangeFeedConfig struct {
	Session     feedSession
	Changes     remoteChanges
	Backoff     time.Duration
	MaxBackoff  time.Duration
	IdleTimeout time.Duration
	Logger      *slog.Logger
}

// ChangeFeed holds one server-sent event stream to the Neoworks API while a
// user is signed in and tells the Syncer when something changed elsewhere. It
// reconnects with backoff and the id of the last event it saw, and asks for a
// sync after every connect, because changes made while it was away only reach
// the library through a pass.
type ChangeFeed struct {
	config  ChangeFeedConfig
	wake    chan struct{}
	cancel  context.CancelFunc
	running sync.WaitGroup

	// lastEventID and failures belong to the loop goroutine.
	lastEventID string
	failures    int
}

// errFeedRefused is a status other than 200 from the stream endpoint.
type errFeedRefused struct {
	status     int
	retryAfter time.Duration
}

func (refused errFeedRefused) Error() string {
	return fmt.Sprintf("neoworks: the change feed answered %d", refused.status)
}

// NewChangeFeed builds a ChangeFeed; Start runs it.
func NewChangeFeed(config ChangeFeedConfig) *ChangeFeed {
	if config.Backoff <= 0 {
		config.Backoff = DefaultFeedBackoff
	}
	if config.MaxBackoff <= 0 {
		config.MaxBackoff = DefaultFeedMaxBackoff
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = DefaultFeedIdleTimeout
	}
	if config.Logger == nil {
		config.Logger = slog.New(slog.DiscardHandler)
	}
	return &ChangeFeed{config: config, wake: make(chan struct{}, 1)}
}

// Start connects while a user is signed in, until ctx ends or Stop is called.
// It must be called once.
func (feed *ChangeFeed) Start(ctx context.Context) {
	ctx, feed.cancel = context.WithCancel(ctx)
	feed.running.Add(1)
	go func() {
		defer feed.running.Done()
		feed.loop(ctx)
	}()
}

// Stop closes the stream and waits for the feed to end.
func (feed *ChangeFeed) Stop() {
	if feed.cancel != nil {
		feed.cancel()
	}
	feed.running.Wait()
}

// Wake makes the feed look at the sign-in state now: it connects after a
// sign-in and closes the stream after a sign-out.
func (feed *ChangeFeed) Wake() {
	select {
	case feed.wake <- struct{}{}:
	default:
	}
}

func (feed *ChangeFeed) loop(ctx context.Context) {
	for ctx.Err() == nil {
		feed.sleep(ctx, feed.attempt(ctx))
	}
}

// attempt runs one connection and returns how long to wait before the next.
func (feed *ChangeFeed) attempt(ctx context.Context) time.Duration {
	if !feed.config.Session.SignedIn() {
		feed.forgetStream()
		return signedOutPoll
	}
	err := feed.stream(ctx)
	if ctx.Err() != nil {
		return 0
	}
	if !feed.config.Session.SignedIn() {
		feed.forgetStream()
		return signedOutPoll
	}
	return feed.delayAfter(err)
}

func (feed *ChangeFeed) forgetStream() {
	feed.lastEventID = ""
	feed.failures = 0
}

func (feed *ChangeFeed) sleep(ctx context.Context, delay time.Duration) {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-feed.wake:
	case <-timer.C:
	}
}

// delayAfter is the wait after a stream ended with err: the base delay after a
// clean end, the server's Retry-After when it named one, otherwise a growing
// delay with jitter.
func (feed *ChangeFeed) delayAfter(err error) time.Duration {
	if err == nil {
		feed.failures = 0
		return feed.config.Backoff
	}
	feed.failures++
	feed.logFailure(err)
	var refused errFeedRefused
	if errors.As(err, &refused) && refused.retryAfter > 0 {
		return refused.retryAfter
	}
	delay := feed.config.Backoff << min(feed.failures-1, 20)
	delay = min(delay, feed.config.MaxBackoff)
	return delay + rand.N(delay/5+1)
}

// logFailure warns about the first failure of a streak and keeps the rest quiet.
func (feed *ChangeFeed) logFailure(err error) {
	if feed.failures == 1 {
		feed.config.Logger.Warn("neoworks: change feed", "error", err)
		return
	}
	feed.config.Logger.Debug("neoworks: change feed", "error", err, "failures", feed.failures)
}

// stream opens the stream and reads it to its end. The error is nil when the
// server closed it cleanly.
func (feed *ChangeFeed) stream(ctx context.Context) error {
	token, err := feed.config.Session.accessToken(ctx)
	if err != nil {
		return err
	}
	streamContext, cancel := context.WithCancel(ctx)
	defer cancel()
	response, err := feed.open(streamContext, token)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	feed.failures = 0
	feed.config.Changes.Request()
	go feed.closeOnSignOut(streamContext, cancel)
	idle := time.AfterFunc(feed.config.IdleTimeout, cancel)
	defer idle.Stop()
	err = feed.readEvents(response.Body, func() { idle.Reset(feed.config.IdleTimeout) })
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func (feed *ChangeFeed) open(ctx context.Context, token string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.config.Session.changesURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("neoworks: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "text/event-stream")
	if feed.lastEventID != "" {
		request.Header.Set("Last-Event-ID", feed.lastEventID)
	}
	response, err := feed.config.Session.streamClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("neoworks: change feed: %w", err)
	}
	if response.StatusCode == http.StatusOK {
		return response, nil
	}
	_ = response.Body.Close()
	return nil, errFeedRefused{status: response.StatusCode, retryAfter: retryAfter(response)}
}

func retryAfter(response *http.Response) time.Duration {
	seconds, err := strconv.Atoi(response.Header.Get("Retry-After"))
	if err != nil || seconds < 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

// closeOnSignOut ends the stream when a wake finds nobody signed in.
func (feed *ChangeFeed) closeOnSignOut(ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-feed.wake:
			if !feed.config.Session.SignedIn() {
				cancel()
				return
			}
		}
	}
}

func (feed *ChangeFeed) readEvents(body io.Reader, onLine func()) error {
	reader := newEventReader(body, onLine)
	for {
		event, err := reader.next()
		if err != nil {
			return err
		}
		feed.handle(event)
	}
}

type changePayload struct {
	Sequence int64 `json:"seq"`
}

// handle passes an event on. An event it cannot read counts as a resync, since
// something changed and the safe answer is to pull.
func (feed *ChangeFeed) handle(event serverEvent) {
	switch event.name {
	case "change":
		feed.handleChange(event)
	case "resync":
		feed.config.Changes.NoteResync()
	}
	if event.hasID {
		feed.lastEventID = event.id
	}
}

func (feed *ChangeFeed) handleChange(event serverEvent) {
	var payload changePayload
	if err := json.Unmarshal([]byte(event.data), &payload); err != nil || payload.Sequence <= 0 {
		feed.config.Logger.Warn("neoworks: change feed event not understood", "data", event.data)
		feed.config.Changes.NoteResync()
		return
	}
	feed.config.Changes.NoteRemoteChange(payload.Sequence)
}

// changesURL is the address of the change stream.
func (session *Session) changesURL() string {
	return session.config.APIURL + changesPath
}

// streamClient is the session's HTTP client without a request timeout, which
// would end a stream that is meant to stay open.
func (session *Session) streamClient() *http.Client {
	client := *session.config.HTTPClient
	client.Timeout = 0
	return &client
}

var (
	_ remoteChanges = (*Syncer)(nil)
	_ feedSession   = (*Session)(nil)
)
