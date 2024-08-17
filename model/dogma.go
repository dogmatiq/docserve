package model

import (
	"context"
	"log/slog"

	"github.com/dogmatiq/configkit"
)

// AppDiscovered is a message that indicates a Dogma application was discovered.
type AppDiscovered struct {
	Repo   Repo
	Module Module
	App    configkit.Application
}

// LogTo logs the message to the given logger.
func (m AppDiscovered) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"dogma application discovered",
		m.Repo.AsLogAttr(),
		m.Module.AsLogAttr(),
		slog.Group(
			"app",
			slog.String("key", m.App.Identity().Key),
			slog.String("name", m.App.Identity().Name),
		),
	)
}
