package repo

import (
	"context"
	"log/slog"
)

// Repo hosts the basic details of a generic repository.
type Repo struct {
	Source string
	ID     string
	Name   string
}

// Found is a message that indicates a repository was found.
type Found interface {
	FoundRepo() Repo
	LogTo(ctx context.Context, logger *slog.Logger)
}

// LogFound logs the message to the given logger.
func LogFound(ctx context.Context, m Found, logger *slog.Logger) {
	r := m.FoundRepo()

	logger.DebugContext(
		ctx,
		"repository found",
		slog.Group(
			"repo",
			slog.String("source", r.Source),
			slog.String("id", r.ID),
			slog.String("name", r.Name),
		),
	)
}

// Lost is a message that indicates a repository is lost, either because it
// has been deleted or is no longer accessible to the browser.
type Lost interface {
	LostRepo() Repo
	LogTo(ctx context.Context, logger *slog.Logger)
}

// LogLost logs the message to the given logger.
func LogLost(ctx context.Context, m Lost, logger *slog.Logger) {
	r := m.LostRepo()

	logger.DebugContext(
		ctx,
		"repository lost",
		slog.Group(
			"repo",
			slog.String("source", r.Source),
			slog.String("id", r.ID),
			slog.String("name", r.Name),
		),
	)
}
