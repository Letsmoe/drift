package library

import (
	"context"
	"fmt"
	"time"
)

// BlockKind is what a recommendation block applies to.
type BlockKind string

// Block kinds. The values are what the recommendation_block table stores.
const (
	BlockTrack  BlockKind = "track"
	BlockArtist BlockKind = "artist"
)

// RecommendationBlock is a track or artist the user asked never to be recommended.
type RecommendationBlock struct {
	Kind BlockKind `json:"kind"`
	// ID is the video ID of a track or the browse ID of an artist.
	ID string `json:"id"`
	// Name is the track title or artist name, for listing the blocks.
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// BlockRecommendation keeps a track or artist out of recommendations and autoplay;
// blocking it again only refreshes its name.
func (library *Library) BlockRecommendation(ctx context.Context, kind BlockKind, id, name string) error {
	if id == "" {
		return fmt.Errorf("library: block %s: empty id", kind)
	}
	_, err := library.db.ExecContext(ctx, `INSERT INTO recommendation_block (kind, id, name, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (kind, id) DO UPDATE SET name = excluded.name`,
		string(kind), id, name, library.timestamp())
	if err != nil {
		return fmt.Errorf("library: block %s %s: %w", kind, id, err)
	}
	return nil
}

// UnblockRecommendation lifts a block; lifting one that does not exist is not an error.
func (library *Library) UnblockRecommendation(ctx context.Context, kind BlockKind, id string) error {
	_, err := library.db.ExecContext(ctx, `DELETE FROM recommendation_block WHERE kind = ? AND id = ?`, string(kind), id)
	if err != nil {
		return fmt.Errorf("library: unblock %s %s: %w", kind, id, err)
	}
	return nil
}

// RecommendationBlocks lists every block, the newest first.
func (library *Library) RecommendationBlocks(ctx context.Context) ([]RecommendationBlock, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT kind, id, name, created_at FROM recommendation_block
		ORDER BY created_at DESC, rowid DESC`)
	if err != nil {
		return nil, fmt.Errorf("library: query recommendation blocks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	blocks := []RecommendationBlock{}
	for rows.Next() {
		block, err := scanBlock(rows.Scan)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("library: read recommendation blocks: %w", err)
	}
	return blocks, nil
}

func scanBlock(scan func(...any) error) (RecommendationBlock, error) {
	var block RecommendationBlock
	var kind, createdAt string
	if err := scan(&kind, &block.ID, &block.Name, &createdAt); err != nil {
		return RecommendationBlock{}, fmt.Errorf("library: scan recommendation block: %w", err)
	}
	block.Kind = BlockKind(kind)
	parsed, err := ParseTime(createdAt)
	if err != nil {
		return RecommendationBlock{}, fmt.Errorf("library: recommendation block time: %w", err)
	}
	block.CreatedAt = parsed
	return block, nil
}
