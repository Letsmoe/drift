package neoworks

import (
	"context"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/letsmoe/drift/core/library"
)

const dayLayout = "2006-01-02"

// Device names this install in the history it writes.
type Device struct {
	ID   string
	Name string
}

type playContent struct {
	VideoID         string `json:"videoId"`
	Context         string `json:"context,omitempty"`
	StartedAt       string `json:"startedAt"`
	ProgressSeconds int    `json:"progressSeconds"`
	Completed       bool   `json:"completed"`
}

type searchContent struct {
	Query      string `json:"query"`
	SearchedAt string `json:"searchedAt"`
}

type dayLogContent struct {
	DeviceID   string          `json:"deviceId"`
	DeviceName string          `json:"deviceName,omitempty"`
	Day        string          `json:"day"`
	Plays      []playContent   `json:"plays,omitempty"`
	Searches   []searchContent `json:"searches,omitempty"`
}

// historyMapping maps plays and searches as one log per device and UTC day,
// keyed by device ID and day. Only the device a log belongs to writes it: this
// install's logs are never applied back, and other installs' logs only add
// plays and searches.
type historyMapping struct {
	library *library.Library
	device  Device
}

func (historyMapping) name() string { return "history" }

func (historyMapping) optionalFields() []string {
	return []string{"deviceName", "plays", "searches"}
}

func (historyMapping) keepsNodes() bool { return true }

func (mapping historyMapping) ownsKey(key string) bool {
	return strings.HasPrefix(key, mapping.device.ID+"|")
}

func dayLogKey(deviceID, day string) string {
	return deviceID + "|" + day
}

func (mapping historyMapping) local(ctx context.Context) ([]entity, error) {
	plays, err := mapping.library.OwnPlays(ctx)
	if err != nil {
		return nil, err
	}
	searches, err := mapping.library.OwnSearches(ctx)
	if err != nil {
		return nil, err
	}
	logs := mapping.groupByDay(plays, searches)
	entities := make([]entity, 0, len(logs))
	for _, day := range slices.Sorted(maps.Keys(logs)) {
		item, err := dayLogEntity(*logs[day], time.Time{})
		if err != nil {
			return nil, err
		}
		entities = append(entities, item)
	}
	return entities, nil
}

// groupByDay files plays and searches under the UTC day they happened on, so
// a log never moves between days when the device changes time zone.
func (mapping historyMapping) groupByDay(plays []library.PlayRecord, searches []library.SearchRecord) map[string]*dayLogContent {
	logs := map[string]*dayLogContent{}
	logOf := func(moment time.Time) *dayLogContent {
		day := moment.UTC().Format(dayLayout)
		if logs[day] == nil {
			logs[day] = &dayLogContent{DeviceID: mapping.device.ID, DeviceName: mapping.device.Name, Day: day}
		}
		return logs[day]
	}
	for _, play := range plays {
		log := logOf(play.StartedAt)
		log.Plays = append(log.Plays, playContent{
			VideoID:         play.VideoID,
			Context:         play.Context,
			StartedAt:       library.FormatTime(play.StartedAt),
			ProgressSeconds: play.ProgressSeconds,
			Completed:       play.Completed,
		})
	}
	for _, search := range searches {
		log := logOf(search.SearchedAt)
		log.Searches = append(log.Searches, searchContent{Query: search.Query, SearchedAt: library.FormatTime(search.SearchedAt)})
	}
	return logs
}

func dayLogEntity(content dayLogContent, updatedAt time.Time) (entity, error) {
	canonical, err := canonicalJSON(content)
	if err != nil {
		return entity{}, err
	}
	return entity{key: dayLogKey(content.DeviceID, content.Day), canonical: canonical, updatedAt: updatedAt, model: content}, nil
}

func (historyMapping) decode(_ string, raw []byte) (entity, error) {
	content, err := decodeModel[dayLogContent](raw)
	if err != nil {
		return entity{}, err
	}
	err = normalizeDayLogTimes(&content)
	if err != nil {
		return entity{}, err
	}
	return dayLogEntity(content, time.Time{})
}

func normalizeDayLogTimes(content *dayLogContent) error {
	var err error
	for index := range content.Plays {
		content.Plays[index].StartedAt, err = normalizeTime(content.Plays[index].StartedAt)
		if err != nil {
			return err
		}
	}
	for index := range content.Searches {
		content.Searches[index].SearchedAt, err = normalizeTime(content.Searches[index].SearchedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mapping historyMapping) apply(ctx context.Context, remote entity) error {
	content := remote.model.(dayLogContent)
	videoIDs := make([]string, 0, len(content.Plays))
	plays := make([]library.PlayRecord, 0, len(content.Plays))
	for _, play := range content.Plays {
		videoIDs = append(videoIDs, play.VideoID)
		plays = append(plays, library.PlayRecord{
			VideoID:         play.VideoID,
			Context:         play.Context,
			StartedAt:       parseNodeTime(play.StartedAt),
			ProgressSeconds: play.ProgressSeconds,
			Completed:       play.Completed,
		})
	}
	err := requireTracks(ctx, mapping.library, videoIDs...)
	if err != nil {
		return err
	}
	err = mapping.library.RestorePlays(ctx, content.DeviceID, plays)
	if err != nil {
		return err
	}
	searches := make([]library.SearchRecord, 0, len(content.Searches))
	for _, search := range content.Searches {
		searches = append(searches, library.SearchRecord{Query: search.Query, SearchedAt: parseNodeTime(search.SearchedAt)})
	}
	return mapping.library.RestoreSearches(ctx, content.DeviceID, searches)
}

// remove keeps what was played: history records facts, and another install
// deleting its log does not undo them here.
func (historyMapping) remove(context.Context, string) error { return nil }
