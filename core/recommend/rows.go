package recommend

import "sort"

// groupRows buckets the best-first items by reason, drops thin and
// YouTube-only rows, orders rows by the mean score of their top three, caps the
// "Because you listened to" rows and interleaves row kinds. Items whose video
// ID is in excluded (the "Made for you" block) never appear, and a track
// appears in one row only.
func groupRows(scored []Item, excluded map[string]bool, rowLimit, perRow int) []Row {
	rows := []Row{}
	for _, row := range bucketByReason(withoutRepeats(scored, excluded)) {
		if row.Reason.Kind != ReasonYouTube && len(row.Items) >= minRowItems {
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(left, right int) bool {
		return rowStrength(rows[left]) > rowStrength(rows[right])
	})
	rows = capRelatedRows(rows)
	return trimRows(interleaveKinds(rows), rowLimit, perRow)
}

// withoutRepeats drops excluded tracks and every later occurrence of a track.
func withoutRepeats(scored []Item, excluded map[string]bool) []Item {
	seen := map[string]bool{}
	kept := make([]Item, 0, len(scored))
	for _, item := range scored {
		videoID := item.Track.VideoID
		if excluded[videoID] || seen[videoID] {
			continue
		}
		seen[videoID] = true
		kept = append(kept, item)
	}
	return kept
}

// capRelatedRows keeps the strongest maxRelatedRows "Because you listened to"
// rows; rows must be sorted strongest first.
func capRelatedRows(rows []Row) []Row {
	kept := make([]Row, 0, len(rows))
	relatedCount := 0
	for _, row := range rows {
		if row.Reason.Kind == ReasonRelated {
			relatedCount++
			if relatedCount > maxRelatedRows {
				continue
			}
		}
		kept = append(kept, row)
	}
	return kept
}

// interleaveKinds deals rows round-robin by reason kind, so rows of one kind
// are not adjacent while another kind is left. Kinds start in order of their
// strongest row and rows keep their strength order within a kind.
func interleaveKinds(rows []Row) []Row {
	kinds := []string{}
	rowsByKind := map[string][]Row{}
	for _, row := range rows {
		kind := row.Reason.Kind
		if _, found := rowsByKind[kind]; !found {
			kinds = append(kinds, kind)
		}
		rowsByKind[kind] = append(rowsByKind[kind], row)
	}
	mixed := make([]Row, 0, len(rows))
	for len(mixed) < len(rows) {
		for _, kind := range kinds {
			queue := rowsByKind[kind]
			if len(queue) == 0 {
				continue
			}
			mixed = append(mixed, queue[0])
			rowsByKind[kind] = queue[1:]
		}
	}
	return mixed
}

// bucketByReason groups items by reason key, in order of each group's first item.
func bucketByReason(scored []Item) []Row {
	indexByKey := map[string]int{}
	rows := []Row{}
	for _, item := range scored {
		key := item.Reason.Key()
		index, found := indexByKey[key]
		if !found {
			index = len(rows)
			indexByKey[key] = index
			rows = append(rows, Row{Key: key, Title: item.Reason.Text, Reason: item.Reason})
		}
		rows[index].Items = append(rows[index].Items, item)
	}
	return rows
}

// rowStrength is the mean score of a row's top three items.
func rowStrength(row Row) float64 {
	count := min(3, len(row.Items))
	sum := 0.0
	for _, item := range row.Items[:count] {
		sum += item.Score
	}
	return sum / float64(count)
}

// trimRows caps the number of rows and the items per row.
func trimRows(rows []Row, rowLimit, perRow int) []Row {
	if len(rows) > rowLimit {
		rows = rows[:max(rowLimit, 0)]
	}
	for index := range rows {
		if len(rows[index].Items) > perRow {
			rows[index].Items = rows[index].Items[:max(perRow, 0)]
		}
	}
	return rows
}
