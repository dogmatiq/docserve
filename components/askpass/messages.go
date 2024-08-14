package askpass

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/google/uuid"
)

// Request is a message that requests credentials for the repository located at
// the given URL. The URL might not contain the complete path to the repository.
type Request struct {
	RequestID uuid.UUID
	RepoURL   *url.URL
	Field     Field
}

// Field is an enumeration of the fields that can be requested by askpass.
type Field int

const (
	// Username requests the username.
	Username Field = iota

	// Password requests the password.
	Password
)

// LogTo logs the message to the given logger.
func (m Request) LogTo(ctx context.Context, logger *slog.Logger) {
	message := "askpass username request"
	if m.Field == Password {
		message = "askpass password request"
	}

	logger.DebugContext(
		ctx,
		message,
		slog.String("request_id", m.RequestID.String()),
		slog.Group(
			"repo",
			slog.String("url", m.RepoURL.String()),
		),
	)
}

// Response is a response to a [Request] message.
type Response struct {
	RequestID uuid.UUID

	RepoSource string
	RepoID     string
	RepoURL    *url.URL

	Field Field
	Value string
}

// LogTo logs the message to the given logger.
func (m Response) LogTo(ctx context.Context, logger *slog.Logger) {
	message := "askpass username response"
	if m.Field == Password {
		message = "askpass password response"
	}

	logger.DebugContext(
		ctx,
		message,
		slog.String("request_id", m.RequestID.String()),
		slog.Group(
			"repo",
			slog.String("source", m.RepoSource),
			slog.String("id", m.RepoID),
			slog.String("url", m.RepoURL.String()),
		),
	)
}
