package neoworks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/letsmoe/drift/core/library"
)

// errTrackWithoutRow means an entity refers to a track that has no Track row
// to point at. The track pass runs first and creates one for every track
// another table mentions, so this only happens when that pass failed.
var errTrackWithoutRow = errors.New("the track has no row yet")

// trackRows translates between a track's video ID, which the library and the
// entities use, and the ID of its Track row, which references in row content
// hold. It is read from the links the track pass keeps; a Track row no link
// names, a duplicate another install created offline, is read when met.
type trackRows struct {
	nodes      Nodes
	rowOfVideo map[string]string
	videoOfRow map[string]string
}

func loadTrackRows(ctx context.Context, store *library.Library, nodes Nodes) (trackRows, error) {
	links, err := store.NodeLinks(ctx, trackMapping{}.name())
	if err != nil {
		return trackRows{}, err
	}
	rows := trackRows{nodes: nodes, rowOfVideo: map[string]string{}, videoOfRow: map[string]string{}}
	for videoID, link := range links {
		rows.rowOfVideo[videoID] = link.NodeID
		rows.videoOfRow[link.NodeID] = videoID
	}
	return rows, nil
}

func (rows trackRows) rowOf(videoID string) (string, error) {
	rowID, found := rows.rowOfVideo[videoID]
	if !found {
		return "", fmt.Errorf("%s: %w", videoID, errTrackWithoutRow)
	}
	return rowID, nil
}

func (rows trackRows) videoOf(rowID string) (string, error) {
	videoID, found := rows.videoOfRow[rowID]
	if found {
		return videoID, nil
	}
	raw, err := rows.nodes.ReadNode(rowID)
	if err != nil {
		return "", fmt.Errorf("track row %s: %w", rowID, errTrackNotSynced)
	}
	content, err := decodeModel[trackContent](raw)
	if err != nil {
		return "", fmt.Errorf("track row %s: %w", rowID, err)
	}
	rows.videoOfRow[rowID] = content.VideoID
	return content.VideoID, nil
}

// referencingMapping is a mapping whose rows refer to Track rows. Its entities
// name tracks by video ID, so hashes and conflict decisions do not depend on
// row IDs; only the row content holds Track row IDs.
type referencingMapping interface {
	// toRow is the content of the entity's row.
	toRow(item entity, tracks trackRows) (map[string]any, error)
	// fromRow turns row content into the content decode reads.
	fromRow(raw []byte, tracks trackRows) ([]byte, error)
}

// contentMap is a typed content as the map libneoworks takes.
func contentMap(model any) (map[string]any, error) {
	encoded, err := canonicalJSON(model)
	if err != nil {
		return nil, err
	}
	var content map[string]any
	err = json.Unmarshal(encoded, &content)
	if err != nil {
		return nil, fmt.Errorf("decode content: %w", err)
	}
	return content, nil
}
