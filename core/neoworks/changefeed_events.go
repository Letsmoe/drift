package neoworks

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// serverEvent is one event of a text/event-stream body.
type serverEvent struct {
	// id is the event's id field; hasID tells an absent id from an empty one.
	id    string
	hasID bool
	name  string
	data  string
}

// eventReader reads server-sent events the way the Neoworks API writes them:
// lines end in LF, a blank line ends an event, lines starting with a colon are
// comments, and events without data are dropped.
type eventReader struct {
	scanner *bufio.Scanner
	// onLine is called for every line read, comments included, so the caller
	// can tell the stream is alive.
	onLine func()
}

func newEventReader(body io.Reader, onLine func()) *eventReader {
	return &eventReader{scanner: bufio.NewScanner(body), onLine: onLine}
}

// next returns the next event with data, or io.EOF when the stream ends.
func (reader *eventReader) next() (serverEvent, error) {
	var event serverEvent
	var dataLines []string
	for reader.scanner.Scan() {
		line := reader.scanner.Text()
		reader.onLine()
		if line != "" {
			dataLines = event.apply(line, dataLines)
			continue
		}
		if len(dataLines) > 0 {
			event.data = strings.Join(dataLines, "\n")
			return event, nil
		}
		event = serverEvent{}
	}
	if err := reader.scanner.Err(); err != nil {
		return serverEvent{}, fmt.Errorf("neoworks: change feed: %w", err)
	}
	return serverEvent{}, io.EOF
}

// apply reads one field line into the event and returns the data lines so far.
func (event *serverEvent) apply(line string, dataLines []string) []string {
	if strings.HasPrefix(line, ":") {
		return dataLines
	}
	field, value, _ := strings.Cut(line, ":")
	value = strings.TrimPrefix(value, " ")
	switch field {
	case "id":
		event.id, event.hasID = value, true
	case "event":
		event.name = value
	case "data":
		dataLines = append(dataLines, value)
	}
	return dataLines
}
