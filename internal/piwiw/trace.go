package piwiw

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const traceTimeLayout = "20060102T150405Z"

type traceSession struct {
	dir string
}

// newTraceSession creates the per-request trace folder if tracing is enabled.
// Returns nil (a valid no-op receiver) when tracing is disabled or the folder
// could not be created, so callers never need to nil-check before use.
func newTraceSession(requestID string) *traceSession {
	if cfg.TraceFolderPath == "" {
		return nil
	}

	name := time.Now().UTC().Format(traceTimeLayout) + "." + requestID
	dir := filepath.Join(cfg.TraceFolderPath, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		logW(requestID, "Failed to create trace folder %s: %v", dir, err)
		return nil
	}
	return &traceSession{dir: dir}
}

func (t *traceSession) writeJSON(requestID, filename string, raw []byte) {
	if t == nil {
		return
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		pretty.Reset()
		pretty.Write(raw)
	}
	if err := os.WriteFile(filepath.Join(t.dir, filename), pretty.Bytes(), 0o644); err != nil {
		logW(requestID, "Failed to write trace file %s: %v", filename, err)
	}
}

// startTraceCleanup periodically removes trace folders older than
// cfg.TraceKeepHours, based on the timestamp encoded in their folder name.
// It runs until stop is closed.
func startTraceCleanup(stop <-chan struct{}) {
	if cfg.TraceFolderPath == "" {
		return
	}

	cleanupTraces()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cleanupTraces()
		case <-stop:
			return
		}
	}
}

func cleanupTraces() {
	entries, err := os.ReadDir(cfg.TraceFolderPath)
	if err != nil {
		logW("-", "Failed to read trace folder %s: %v", cfg.TraceFolderPath, err)
		return
	}

	cutoff := time.Now().UTC().Add(-time.Duration(cfg.TraceKeepHours) * time.Hour)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		dot := strings.Index(name, ".")
		if dot <= 0 {
			continue
		}
		ts, err := time.Parse(traceTimeLayout, name[:dot])
		if err != nil {
			continue
		}
		if ts.Before(cutoff) {
			path := filepath.Join(cfg.TraceFolderPath, name)
			if err := os.RemoveAll(path); err != nil {
				logW("-", "Failed to remove expired trace folder %s: %v", path, err)
			}
		}
	}
}
