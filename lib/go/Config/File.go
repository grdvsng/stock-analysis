package Config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ARCHIVE_TIME_FORMAT = "2006-01-02_15-04-05"

func Validate(raw []byte, near string) (*Config, error) {
	if !json.Valid(raw) {
		return nil, InvalidValue("config", "not a json")
	}

	temporary, e := os.CreateTemp(filepath.Dir(near), ".config-*.json")

	if e != nil {
		return nil, e
	}

	defer os.Remove(temporary.Name())

	if _, e := temporary.Write(raw); e != nil {
		temporary.Close()

		return nil, e
	}

	temporary.Close()

	return Load(temporary.Name())
}

func Archive(path string, now time.Time) (string, error) {
	raw, e := os.ReadFile(path)

	if e != nil {
		return "", e
	}

	extension := filepath.Ext(path)
	archive := fmt.Sprintf("%s.%s%s", strings.TrimSuffix(path, extension), now.Format(ARCHIVE_TIME_FORMAT), extension)

	return archive, os.WriteFile(archive, raw, 0o644)
}

func Replace(path string, raw []byte) error {
	temporary := path + ".tmp"

	if e := os.WriteFile(temporary, raw, 0o644); e != nil {
		return e
	}

	return os.Rename(temporary, path)
}

func Restore(path, archive string) error {
	raw, e := os.ReadFile(archive)

	if e != nil {
		return e
	}

	return Replace(path, raw)
}
