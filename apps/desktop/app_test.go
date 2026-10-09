package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateLegacyDirectoryMovesOldDirectory(t *testing.T) {
	root := t.TempDir()
	legacyDirectory := filepath.Join(root, "otter-music")
	directory := filepath.Join(root, "drift")
	writeTestFile(t, filepath.Join(legacyDirectory, "library.db"), "legacy")

	migrateLegacyDirectory(discardLogger(), legacyDirectory, directory)

	content, err := os.ReadFile(filepath.Join(directory, "library.db"))
	if err != nil {
		t.Fatalf("read migrated library: %v", err)
	}
	if string(content) != "legacy" {
		t.Errorf("migrated content = %q, want legacy", content)
	}
	_, err = os.Stat(legacyDirectory)
	if !os.IsNotExist(err) {
		t.Errorf("legacy directory still exists (err = %v)", err)
	}
}

func TestMigrateLegacyDirectoryKeepsExistingDirectory(t *testing.T) {
	root := t.TempDir()
	legacyDirectory := filepath.Join(root, "otter-music")
	directory := filepath.Join(root, "drift")
	writeTestFile(t, filepath.Join(legacyDirectory, "library.db"), "legacy")
	writeTestFile(t, filepath.Join(directory, "library.db"), "current")

	migrateLegacyDirectory(discardLogger(), legacyDirectory, directory)

	content, err := os.ReadFile(filepath.Join(directory, "library.db"))
	if err != nil {
		t.Fatalf("read library: %v", err)
	}
	if string(content) != "current" {
		t.Errorf("content = %q, want current", content)
	}
	_, err = os.Stat(legacyDirectory)
	if err != nil {
		t.Errorf("legacy directory should stay untouched: %v", err)
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	err := os.MkdirAll(filepath.Dir(path), 0o700)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}
