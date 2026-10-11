package neoworks

import (
	"context"
	"fmt"
	"time"

	"github.com/letsmoe/drift/core/library"
)

type blockContent struct {
	Kind      string `json:"kind"`
	TargetID  string `json:"targetId"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

// blockMapping maps recommendation blocks, keyed by kind and target ID.
type blockMapping struct {
	library *library.Library
}

func (blockMapping) name() string { return "blocks" }

func (blockMapping) optionalFields() []string { return nil }

func (blockMapping) keepsNodes() bool { return false }

func (blockMapping) ownsKey(string) bool { return false }

func (mapping blockMapping) local(ctx context.Context) ([]entity, error) {
	blocks, err := mapping.library.RecommendationBlocks(ctx)
	if err != nil {
		return nil, err
	}
	entities := make([]entity, 0, len(blocks))
	for _, block := range blocks {
		content := blockContent{
			Kind:      string(block.Kind),
			TargetID:  block.ID,
			Name:      block.Name,
			CreatedAt: library.FormatTime(block.CreatedAt),
		}
		item, err := blockEntity(content, block.CreatedAt)
		if err != nil {
			return nil, err
		}
		entities = append(entities, item)
	}
	return entities, nil
}

func blockEntity(content blockContent, createdAt time.Time) (entity, error) {
	canonical, err := canonicalJSON(content)
	if err != nil {
		return entity{}, err
	}
	return entity{key: savedKey(content.Kind, content.TargetID), canonical: canonical, updatedAt: createdAt, model: content}, nil
}

func (blockMapping) decode(_ string, raw []byte) (entity, error) {
	content, err := decodeModel[blockContent](raw)
	if err != nil {
		return entity{}, err
	}
	content.CreatedAt, err = normalizeTime(content.CreatedAt)
	if err != nil {
		return entity{}, err
	}
	return blockEntity(content, parseNodeTime(content.CreatedAt))
}

func (mapping blockMapping) apply(ctx context.Context, remote entity) error {
	content := remote.model.(blockContent)
	kind := library.BlockKind(content.Kind)
	switch kind {
	case library.BlockTrack, library.BlockArtist, library.BlockPodcast:
	default:
		return fmt.Errorf("unknown block kind %q", content.Kind)
	}
	return mapping.library.RestoreBlock(ctx, library.RecommendationBlock{
		Kind:      kind,
		ID:        content.TargetID,
		Name:      content.Name,
		CreatedAt: remote.updatedAt,
	})
}

func (mapping blockMapping) remove(ctx context.Context, key string) error {
	kind, targetID, err := splitKey(key)
	if err != nil {
		return err
	}
	return mapping.library.UnblockRecommendation(ctx, library.BlockKind(kind), targetID)
}
