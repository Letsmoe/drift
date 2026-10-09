package lyrics

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/letsmoe/drift/core/music"
)

// lrcTimestamp matches one leading "[mm:ss]", "[mm:ss.x]", "[mm:ss.xx]" or "[mm:ss.xxx]" tag.
var lrcTimestamp = regexp.MustCompile(`^\[(\d+):(\d{1,2})(?:[.:](\d{1,3}))?\]`)

// ParseLRC reads LRC lyrics into lines sorted by start time. Metadata tags
// ([ar:...], [offset:...]) and untimed lines are left out; a line with several
// timestamps appears once per timestamp.
func ParseLRC(text string) []music.LyricLine {
	var lines []music.LyricLine
	for _, rawLine := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		lines = append(lines, parseLRCLine(strings.TrimSpace(rawLine))...)
	}
	sort.SliceStable(lines, func(left, right int) bool {
		return lines[left].StartMs < lines[right].StartMs
	})
	return lines
}

func parseLRCLine(rawLine string) []music.LyricLine {
	var startTimes []int64
	rest := rawLine
	for {
		match := lrcTimestamp.FindStringSubmatch(rest)
		if match == nil {
			break
		}
		startTimes = append(startTimes, timestampMs(match))
		rest = rest[len(match[0]):]
	}

	text := strings.TrimSpace(rest)
	lines := make([]music.LyricLine, 0, len(startTimes))
	for _, startMs := range startTimes {
		lines = append(lines, music.LyricLine{Text: text, StartMs: startMs})
	}
	return lines
}

// timestampMs converts a lrcTimestamp match; the fraction is read as decimal digits.
func timestampMs(match []string) int64 {
	minutes, _ := strconv.ParseInt(match[1], 10, 64)
	seconds, _ := strconv.ParseInt(match[2], 10, 64)
	fraction := match[3]
	var fractionMs int64
	if fraction != "" {
		fractionMs, _ = strconv.ParseInt((fraction + "00")[:3], 10, 64)
	}
	return (minutes*60+seconds)*1000 + fractionMs
}
