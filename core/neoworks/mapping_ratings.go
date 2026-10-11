package neoworks

import (
	"context"
	"fmt"
	"time"

	"github.com/letsmoe/drift/core/library"
)

type ratingContent struct {
	VideoID   string `json:"videoId"`
	Opinion   string `json:"opinion"`
	UpdatedAt string `json:"updatedAt"`
}

const (
	opinionLike    = "like"
	opinionDislike = "dislike"
)

// ratingMapping maps the liked and disliked tracks, keyed by video ID.
type ratingMapping struct {
	library *library.Library
}

func (ratingMapping) name() string { return "ratings" }

func (ratingMapping) optionalFields() []string { return nil }

func (ratingMapping) keepsNodes() bool { return false }

func (ratingMapping) ownsKey(string) bool { return false }

func (mapping ratingMapping) local(ctx context.Context) ([]entity, error) {
	records, err := mapping.library.RatingRecords(ctx)
	if err != nil {
		return nil, err
	}
	entities := make([]entity, 0, len(records))
	for _, record := range records {
		content := ratingContent{
			VideoID:   record.VideoID,
			Opinion:   opinionOf(record.Rating),
			UpdatedAt: library.FormatTime(record.UpdatedAt),
		}
		item, err := ratingEntity(content, record.UpdatedAt)
		if err != nil {
			return nil, err
		}
		entities = append(entities, item)
	}
	return entities, nil
}

func opinionOf(rating library.Rating) string {
	if rating == library.RatingDislike {
		return opinionDislike
	}
	return opinionLike
}

func ratingOf(opinion string) (library.Rating, error) {
	switch opinion {
	case opinionLike:
		return library.RatingLike, nil
	case opinionDislike:
		return library.RatingDislike, nil
	}
	return library.RatingNone, fmt.Errorf("unknown opinion %q", opinion)
}

func ratingEntity(content ratingContent, updatedAt time.Time) (entity, error) {
	canonical, err := canonicalJSON(content)
	if err != nil {
		return entity{}, err
	}
	return entity{key: content.VideoID, canonical: canonical, updatedAt: updatedAt, model: content}, nil
}

func (ratingMapping) decode(_ string, raw []byte) (entity, error) {
	content, err := decodeModel[ratingContent](raw)
	if err != nil {
		return entity{}, err
	}
	content.UpdatedAt, err = normalizeTime(content.UpdatedAt)
	if err != nil {
		return entity{}, err
	}
	return ratingEntity(content, parseNodeTime(content.UpdatedAt))
}

func (mapping ratingMapping) apply(ctx context.Context, remote entity) error {
	content := remote.model.(ratingContent)
	rating, err := ratingOf(content.Opinion)
	if err != nil {
		return err
	}
	err = requireTracks(ctx, mapping.library, content.VideoID)
	if err != nil {
		return err
	}
	return mapping.library.RestoreRating(ctx, library.RatingRecord{
		VideoID:   content.VideoID,
		Rating:    rating,
		UpdatedAt: remote.updatedAt,
	})
}

func (mapping ratingMapping) remove(ctx context.Context, key string) error {
	return mapping.library.RestoreRating(ctx, library.RatingRecord{VideoID: key, Rating: library.RatingNone})
}
