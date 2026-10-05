package music

// Lyrics are the words of a song. Plain lyrics carry no timing: Synced is
// false and every StartMs is 0. Synced lyrics set StartMs per line.
type Lyrics struct {
	Lines  []LyricLine `json:"lines"`
	Source string      `json:"source"`
	Synced bool        `json:"synced"`
}

// LyricLine is one line of lyrics. Blank lines separate stanzas and have empty Text.
type LyricLine struct {
	Text    string `json:"text"`
	StartMs int64  `json:"startMs"`
}
