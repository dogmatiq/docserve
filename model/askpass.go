package model

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/google/uuid"
)

// CredentialType is an enumeration of the types of credentials that can be
// requested.
type CredentialType string

const (
	// UsernameCredentialType indicates that relevant credential is the username used to
	// authenticate with the repository.
	UsernameCredentialType CredentialType = "username"
	// PasswordCredentialType indicates that relevant credential is the password used to
	// authenticate with the repository.
	PasswordCredentialType CredentialType = "password"
)

// IsSensitive returns true if the credential is sensitive and should therefore
// not be displayed in plain-text.
func (c CredentialType) IsSensitive() bool {
	return c != UsernameCredentialType
}

// Validate returns an error if the credential is not valid.
func (c CredentialType) Validate() error {
	switch c {
	case UsernameCredentialType, PasswordCredentialType:
		return nil
	default:
		return fmt.Errorf("invalid credential: %q", c)
	}
}

// CredentialRequest is a message that requests a credential for the repository
// located at the given URL.
type CredentialRequest struct {
	RequestID      uuid.UUID
	URL            *url.URL
	CredentialType CredentialType
}

// LogTo logs the message to the given logger.
func (m CredentialRequest) LogTo(ctx context.Context, logger *slog.Logger) {
	logger.DebugContext(
		ctx,
		"repo credential requested",
		slog.String("request_id", m.RequestID.String()),
		slog.String("credential", string(m.CredentialType)),
		slog.String("url", m.URL.String()),
	)
}

// CredentialResponse is a response to a [CredentialRequest] message.
type CredentialResponse struct {
	RequestID uuid.UUID

	Repo           Repo
	CredentialType CredentialType
	Value          string
}

// LogTo logs the message to the given logger.
func (m CredentialResponse) LogTo(ctx context.Context, logger *slog.Logger) {
	attrs := []any{
		slog.String("request_id", m.RequestID.String()),
		slog.String("credential", string(m.CredentialType)),
	}

	if !m.CredentialType.IsSensitive() {
		attrs = append(attrs, slog.String("value", m.Value))
	}

	attrs = append(
		attrs,
		m.Repo.AsLogAttr(),
	)

	logger.DebugContext(
		ctx,
		"repo username provided",
		attrs...,
	)
}
