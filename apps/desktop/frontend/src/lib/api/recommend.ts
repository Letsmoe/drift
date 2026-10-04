import { RecommendedFeed, RecommendedRows } from '#lib/wailsjs/go/main/App.js';
import type { recommend } from '#lib/wailsjs/go/models.ts';

export type RecommendedItem = recommend.Item;
export type RecommendedRow = recommend.Row;
export type RecommendationReason = recommend.Reason;

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
