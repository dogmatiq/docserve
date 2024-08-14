package messages

import (
	"context"
	"log/slog"
)

// ModuleDiscovered is a message that indicates a Go module was found at a specific
// version.
type ModuleDiscovered struct {
	RepoSource    string
	RepoID        string
	ModulePath    string
	ModuleVersion string
}

// LogTo logs the message to the given logger.
func (m ModuleDiscovered) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"module found",
		slog.Group(
			"repo",
			slog.String("source", m.RepoSource),
			slog.String("id", m.RepoID),
		),
		slog.Group(
			"module",
			slog.String("path", m.ModulePath),
			slog.String("version", m.ModuleVersion),
		),
	)
}

// ModuleDownloaded is a message that indicates a Go module was downloaded
// into the module cache.
type ModuleDownloaded struct {
	RepoSource    string
	RepoID        string
	ModulePath    string
	ModuleVersion string
	ModuleDir     string
}

// LogTo logs the message to the given logger.
func (m ModuleDownloaded) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"module downloaded",
		slog.Group(
			"repo",
			slog.String("source", m.RepoSource),
			slog.String("id", m.RepoID),
		),
		slog.Group(
			"module",
			slog.String("path", m.ModulePath),
			slog.String("version", m.ModuleVersion),
		),
	)
}
