import {
	BlockRecommendation,
	RecommendationBlocks,
	RecommendedFeed,
	RecommendedRows,
	UnblockRecommendation
} from '#lib/wailsjs/go/main/App.js';
import type { library, recommend } from '#lib/wailsjs/go/models.ts';

export type RecommendedItem = recommend.Item;
export type RecommendedRow = recommend.Row;
export type RecommendationReason = recommend.Reason;
export type RecommendationBlock = library.RecommendationBlock;
/** What a "Don't recommend" choice applies to: a track (video ID) or an artist (browse ID). */
export type BlockKind = 'track' | 'artist';

export function recommendedRows(rowLimit: number, perRow: number): Promise<RecommendedRow[]> {
	return RecommendedRows(rowLimit, perRow);
}

/** One page of the ranked feed; `refresh` forces a re-rank. */
export function recommendedFeed(
	limit: number,
	offset: number,
	refresh: boolean
): Promise<RecommendedItem[]> {
	return RecommendedFeed(limit, offset, refresh);
}

/** Keeps a track or an artist out of recommendations and autoplay. */
export function blockRecommendation(kind: BlockKind, id: string, name: string): Promise<void> {
	return BlockRecommendation(kind, id, name);
}

export function unblockRecommendation(kind: BlockKind, id: string): Promise<void> {
	return UnblockRecommendation(kind, id);
}

export function recommendationBlocks(): Promise<RecommendationBlock[]> {
	return RecommendationBlocks();
}
