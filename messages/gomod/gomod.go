package gomod

import (
	"context"
	"log/slog"

	"github.com/dogmatiq/browser/messages/repo"
)

// ModuleDiscovered is a message that indicates a Go module was found at a specific
// version.
type ModuleDiscovered struct {
	Repo          repo.Repo
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
			slog.String("source", m.Repo.Source),
			slog.String("id", m.Repo.ID),
			slog.String("name", m.Repo.Name),
		),
		slog.Group(
			"module",
			slog.String("path", m.ModulePath),
			slog.String("version", m.ModuleVersion),
		),
	)
}

// ModuleAvailableOnDisk is a message that indicates a Go module was downloaded
// into the module cache.
type ModuleAvailableOnDisk struct {
	Repo          repo.Repo
	ModulePath    string
	ModuleVersion string
	ModuleDir     string
}

// LogTo logs the message to the given logger.
func (m ModuleAvailableOnDisk) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"module available on disk",
		slog.Group(
			"repo",
			slog.String("source", m.Repo.Source),
			slog.String("id", m.Repo.ID),
			slog.String("name", m.Repo.Name),
		),
		slog.Group(
			"module",
			slog.String("path", m.ModulePath),
			slog.String("version", m.ModuleVersion),
		),
	)
}
