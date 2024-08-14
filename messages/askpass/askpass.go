package askpass

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/google/uuid"
)

// Credential is an enumeration of the types of credentials that can be
// requested.
type Credential string

const (
	// Username indicates that relevant credential is the username used to
	// authenticate with the repository.
	Username Credential = "username"
	// Password indicates that relevant credential is the password used to
	// authenticate with the repository.
	Password Credential = "password"
)

// IsSensitive returns true if the credential is sensitive and should therefore
// not be displayed in plain-text.
func (c Credential) IsSensitive() bool {
	return c != Username
}

// Validate returns an error if the credential is not valid.
func (c Credential) Validate() error {
	switch c {
	case Username, Password:
		return nil
	default:
		return fmt.Errorf("invalid credential: %q", c)
	}
}

// CredentialRequest is a message that requests a credential for the repository
// located at the given URL.
type CredentialRequest struct {
	RequestID  uuid.UUID
	RepoURL    *url.URL
	Credential Credential
}

// LogTo logs the message to the given logger.
func (m CredentialRequest) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"repo credential requested",
		slog.String("request_id", m.RequestID.String()),
		slog.String("credential", string(m.Credential)),
		slog.Group(
			"repo",
			slog.String("url", m.RepoURL.String()),
		),
	)
}

// CredentialResponse is a response to a [CredentialRequest] message.
type CredentialResponse struct {
	RequestID uuid.UUID

	RepoSource string
	RepoID     string
	RepoURL    *url.URL

	Credential Credential
	Value      string
}

// LogTo logs the message to the given logger.
func (m CredentialResponse) LogTo(ctx context.Context, logger *slog.Logger) {
	attrs := []any{
		slog.String("request_id", m.RequestID.String()),
		slog.String("credential", string(m.Credential)),
	}

	if !m.Credential.IsSensitive() {
		attrs = append(attrs, slog.String("value", m.Value))
	}

	attrs = append(
		attrs,
		slog.Group(
			"repo",
			slog.String("source", m.RepoSource),
			slog.String("id", m.RepoID),
			slog.String("url", m.RepoURL.String()),
		),
	)

	logger.DebugContext(
		ctx,
		"repo username provided",
		attrs...,
	)
}
