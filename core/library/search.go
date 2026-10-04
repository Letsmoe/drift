package library

import (
	"context"
	"fmt"
	"strings"
)

// RecordSearch remembers a search query; blank queries are skipped.
func (library *Library) RecordSearch(ctx context.Context, query string) error {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return nil
	}
	_, err := library.db.ExecContext(ctx, `INSERT INTO search_history (query, normalized_query, created_at) VALUES (?, ?, ?)`,
		strings.TrimSpace(query), normalized, library.timestamp())
	if err != nil {
		return fmt.Errorf("library: record search: %w", err)
	}
	return nil
}

// normalizeQuery lowercases and collapses whitespace so near-identical searches match.
func normalizeQuery(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(query)), " ")
}

// RecentSearches returns distinct queries (by normalized form), newest first,
// each as it was last typed.
func (library *Library) RecentSearches(ctx context.Context, limit int) ([]string, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT query FROM search_history
		WHERE id IN (SELECT MAX(id) FROM search_history GROUP BY normalized_query)
		ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("library: query searches: %w", err)
	}
	defer func() { _ = rows.Close() }()
	queries := []string{}
	for rows.Next() {
		var query string
		err = rows.Scan(&query)
		if err != nil {
			return nil, fmt.Errorf("library: scan search: %w", err)
		}
		queries = append(queries, query)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read searches: %w", err)
	}
	return queries, nil
}

// ClearSearches forgets all search history.
func (library *Library) ClearSearches(ctx context.Context) error {
	_, err := library.db.ExecContext(ctx, `DELETE FROM search_history`)
	if err != nil {
		return fmt.Errorf("library: clear searches: %w", err)
	}
	return nil
}
