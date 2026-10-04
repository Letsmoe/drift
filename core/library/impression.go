package library

import (
	"context"
	"database/sql"
	"fmt"
)

// maxImpressionsPerCall bounds one RecordImpressions batch.
const maxImpressionsPerCall = 500

// Impression is a recommendation item that was shown, and maybe clicked.
type Impression struct {
	VideoID string `json:"videoId"`
	// Source is what produced the recommendation, e.g. "radio" or "home".
	Source string `json:"source"`
	// Surface is where it was shown, e.g. "home" or "autoplay".
	Surface  string `json:"surface"`
	Position *int   `json:"position"`
	Clicked  bool   `json:"clicked"`
}

// RecordImpressions stores up to 500 impressions in one transaction; any
// beyond that are dropped.
func (library *Library) RecordImpressions(ctx context.Context, impressions []Impression) error {
	if len(impressions) > maxImpressionsPerCall {
		impressions = impressions[:maxImpressionsPerCall]
	}
	if len(impressions) == 0 {
		return nil
	}
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		shownAt := library.timestamp()
		for _, impression := range impressions {
			_, err := tx.ExecContext(ctx, `INSERT INTO impression (video_id, source, surface, position, clicked, shown_at)
				VALUES (?, ?, ?, ?, ?, ?)`,
				impression.VideoID, impression.Source, impression.Surface, impression.Position, impression.Clicked, shownAt)
			if err != nil {
				return fmt.Errorf("record impression of %s: %w", impression.VideoID, err)
			}
		}
		return nil
	})
}
