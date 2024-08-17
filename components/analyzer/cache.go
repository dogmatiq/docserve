package analyzer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dogmatiq/browser/components/analyzer/internal/analyzerpb"
	"github.com/dogmatiq/configkit"
	"google.golang.org/protobuf/proto"
)

// Cache is an interface for storing and retrieving the results of static
// analysis.
type Cache interface {
	Load(ctx context.Context, path, version string) (CacheEntry, bool, error)
	Save(ctx context.Context, path, version string, entry CacheEntry) error
}

// CacheEntry is a cached result of static analysis.
type CacheEntry struct {
	Apps []configkit.Application
}

// NoopCache is a [Cache] implementation that never stores anything.
type NoopCache struct{}

// Load returns false, indicating no entry was found.
func (NoopCache) Load(context.Context, string, string) (CacheEntry, bool, error) {
	return CacheEntry{}, false, nil
}

// Save does nothing.
func (NoopCache) Save(context.Context, string, string, CacheEntry) error {
	return nil
}

// DiskCache is a [Cache] implementation that stores entries on disk.
type DiskCache struct {
	Dir string
}

// Load loads a cache entry from disk.
func (c DiskCache) Load(_ context.Context, path, version string) (CacheEntry, bool, error) {
	file := filepath.Join(c.Dir, path, version)

	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return CacheEntry{}, false, nil
		}
		return CacheEntry{}, false, fmt.Errorf("unable to read cache file: %w", err)
	}

	if len(data) < sha256.Size {
		return CacheEntry{}, false, fmt.Errorf("cache file is smaller than expected")
	}

	checksum := data[:sha256.Size]
	payload := data[sha256.Size:]
	expected := sha256.Sum256(payload)

	if !bytes.Equal(checksum, expected[:]) {
		return CacheEntry{}, false, fmt.Errorf("cache file checksum does not match")
	}

	e := &analyzerpb.CacheEntry{}
	if err := proto.Unmarshal(payload, e); err != nil {
		return CacheEntry{}, false, fmt.Errorf("unable to unmarshal cache entry: %w", err)
	}

	var entry CacheEntry
	for _, a := range e.Applications {
		app, err := configkit.FromProto(a)
		if err != nil {
			return CacheEntry{}, false, fmt.Errorf("unable to unmarshal application: %w", err)
		}
		entry.Apps = append(entry.Apps, app)
	}

	return entry, true, nil
}

// Save saves a cache entry to disk.
func (c DiskCache) Save(_ context.Context, path, version string, entry CacheEntry) error {
	e := &analyzerpb.CacheEntry{}

	for _, app := range entry.Apps {
		a, err := configkit.ToProto(app)
		if err != nil {
			return fmt.Errorf("unable to marshal %s application: %w", app.Identity(), err)
		}
		e.Applications = append(e.Applications, a)
	}

	payload, err := proto.Marshal(e)
	if err != nil {
		return err
	}
	checksum := sha256.Sum256(payload)
	data := append(checksum[:], payload...)

	dir := filepath.Join(c.Dir, path)
	file := filepath.Join(dir, version)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("unable to create cache directory: %w", err)
	}

	if err := os.WriteFile(file, data, 0600); err != nil {
		return fmt.Errorf("unable to write cache file: %w", err)
	}

	return nil
}
