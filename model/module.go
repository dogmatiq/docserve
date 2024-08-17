package model

import (
	"context"
	"log/slog"
)

// Module hosts the basic details of a Go module.
type Module struct {
	Path    string
	Version string
}

// AsLogAttr returns the structured log attributes for the module.
func (m Module) AsLogAttr() slog.Attr {
	return slog.Group(
		"module",
		slog.String("path", m.Path),
		slog.String("version", m.Version),
	)
}

// ModuleDiscovered is a message that indicates a Go module was found at a specific
// version.
type ModuleDiscovered struct {
	Repo   Repo
	Module Module
}

// LogTo logs the message to the given logger.
func (m ModuleDiscovered) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"module found",
		m.Repo.AsLogAttr(),
		m.Module.AsLogAttr(),
	)
}

// ModuleAvailableOnDisk is a message that indicates a Go module was downloaded
// into the module cache.
type ModuleAvailableOnDisk struct {
	Repo   Repo
	Module Module
	Dir    string
}

// LogTo logs the message to the given logger.
func (m ModuleAvailableOnDisk) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"module available on disk",
		m.Repo.AsLogAttr(),
		m.Module.AsLogAttr(),
	)
}
