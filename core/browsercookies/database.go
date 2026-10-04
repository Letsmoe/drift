package browsercookies

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

// databaseSidecars are the files next to a cookie database that hold changes
// the browser has not merged into it yet.
var databaseSidecars = []string{"", "-wal", "-journal"}

// openDatabaseCopy opens a private copy of a cookie database: browsers lock
// the original while they run. The copy lives in a fresh temp directory that
// the returned cleanup deletes.
func openDatabaseCopy(tempParent, databasePath string) (*sql.DB, func(), error) {
	directory, err := os.MkdirTemp(tempParent, "otter-cookies-")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(directory) }
	copyPath := filepath.Join(directory, "cookies.sqlite")
	err = copyDatabaseFiles(databasePath, copyPath)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	db, err := sql.Open("sqlite", copyPath)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("open cookie database copy: %w", err)
	}
	return db, func() { _ = db.Close(); cleanup() }, nil
}

func copyDatabaseFiles(source, destination string) error {
	for _, suffix := range databaseSidecars {
		err := copyFile(source+suffix, destination+suffix)
		if errors.Is(err, fs.ErrNotExist) && suffix != "" {
			continue
		}
		if err != nil {
			return fmt.Errorf("copy cookie database: %w", err)
		}
	}
	return nil
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open %s: %w", filepath.Base(source), err)
	}
	defer func() { _ = input.Close() }()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(destination), err)
	}
	_, err = io.Copy(output, input)
	closeErr := output.Close()
	if err != nil {
		return fmt.Errorf("copy %s: %w", filepath.Base(source), err)
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", filepath.Base(destination), closeErr)
	}
	return nil
}
