package repo

import (
	"context"
	"log/slog"
)

// Found is a message that indicates a repository was found.
type Found struct {
	RepoSource string
	RepoID     string
	RepoName   string
}

// LogTo logs the message to the given logger.
func (m Found) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"repository found",
		slog.String("repo.source", m.RepoSource),
		slog.String("repo.id", m.RepoID),
		slog.String("repo.name", m.RepoName),
	)
}

// Lost is a message that indicates a repository is lost, either because it
// has been deleted or is no longer accessible to the browser.
type Lost struct {
	RepoSource string
	RepoID     string
}

// LogTo logs the message to the given logger.
func (m Lost) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"repository lost",
		slog.String("repo.source", m.RepoSource),
		slog.String("repo.id", m.RepoID),
	)
}
