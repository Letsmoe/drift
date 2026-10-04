package recommend

import "sort"

// groupRows buckets the best-first items by reason, drops thin and
// YouTube-only rows, and orders rows by the mean score of their top three.
func groupRows(scored []Item, rowLimit, perRow int) []Row {
	rows := []Row{}
	for _, row := range bucketByReason(scored) {
		if row.Reason.Kind != ReasonYouTube && len(row.Items) >= minRowItems {
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(left, right int) bool {
		return rowStrength(rows[left]) > rowStrength(rows[right])
	})
	return trimRows(rows, rowLimit, perRow)
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
