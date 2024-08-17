package model

import (
	"context"
	"log/slog"
)

// Repo hosts the basic details of a generic source repository.
type Repo struct {
	Source string
	ID     string
	Name   string
}

// AsLogAttr returns the structured log attributes for the repository.
func (r Repo) AsLogAttr() slog.Attr {
	return slog.Group(
		"repo",
		slog.String("source", r.Source),
		slog.String("id", r.ID),
		slog.String("name", r.Name),
	)
}

// RepoFound is a message that indicates a repository was found.
type RepoFound interface {
	FoundRepo() Repo
	LogTo(ctx context.Context, logger *slog.Logger)
}

// RepoLost is a message that indicates a repository is lost, either because it
// has been deleted or is no longer accessible to the browser.
type RepoLost interface {
	LostRepo() Repo
	LogTo(ctx context.Context, logger *slog.Logger)
}

// LogRepoFound logs the message to the given logger.
func LogRepoFound(ctx context.Context, m RepoFound, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"repository found",
		m.FoundRepo().AsLogAttr(),
	)
}

// LogRepoLost logs the message to the given logger.
func LogRepoLost(ctx context.Context, m RepoLost, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"repository lost",
		m.LostRepo().AsLogAttr(),
	)
}
