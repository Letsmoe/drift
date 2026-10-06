//go:build live

package lyrics

import "testing"

func TestLiveLRCLIB(t *testing.T) {
	lyrics, found, err := NewLRCLIB(nil, "").Find(t.Context(), faded)
	if err != nil {
		t.Fatal(err)
	}
	if !found || !lyrics.Synced || len(lyrics.Lines) < 10 {
		t.Fatalf("found %v, synced %v, %d lines", found, lyrics.Synced, len(lyrics.Lines))
	}
	t.Logf("%d synced lines, first %q at %d ms", len(lyrics.Lines), lyrics.Lines[1].Text, lyrics.Lines[1].StartMs)
}
