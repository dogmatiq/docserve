package github

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"github.com/dogmatiq/browser/integrations/github/internal/githubapi"
	"github.com/dogmatiq/browser/messages/askpass"
	"github.com/dogmatiq/browser/messages/repo"
	"github.com/dogmatiq/minibus"
	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

// AskpassServer responds to requests for repository credentials.
type AskpassServer struct {
	Client *githubapi.AppClient

	entriesByRepoID   map[string]*askpassEntry
	entriesByRepoName map[string]*askpassEntry
}

type askpassEntry struct {
	Repo        *github.Repository
	TokenSource oauth2.TokenSource
}

// Run starts the server.
func (s *AskpassServer) Run(ctx context.Context) error {
	minibus.Subscribe[repo.Found](ctx)
	minibus.Subscribe[repo.Lost](ctx)
	minibus.Subscribe[askpass.CredentialRequest](ctx)
	minibus.Ready(ctx)

	s.entriesByRepoID = map[string]*askpassEntry{}
	s.entriesByRepoName = map[string]*askpassEntry{}

	for m := range minibus.Inbox(ctx) {
		if err := s.handleMessage(ctx, m); err != nil {
			return err
		}
	}

	return nil
}

func (s *AskpassServer) handleMessage(
	ctx context.Context,
	m any,
) error {
	switch m := m.(type) {
	case repoFound:
		s.handleRepoFound(m)
	case repoLost:
		s.handleRepoLost(m)
	case askpass.CredentialRequest:
		return s.handleAskpassRequest(ctx, m)
	default:
		panic("unexpected message type")
	}

	return nil
}

func (s *AskpassServer) handleRepoFound(m repoFound) {
	if m.AppClientID != s.Client.ID {
		return
	}

	entry := &askpassEntry{
		Repo:        m.GitHubRepo,
		TokenSource: s.Client.InstallationTokenSource(m.InstallationID),
	}

	s.entriesByRepoID[m.Repo.ID] = entry
	s.entriesByRepoName[m.Repo.Name] = entry
}

func (s *AskpassServer) handleRepoLost(m repoLost) {
	if m.AppClientID != s.Client.ID {
		return
	}

	delete(s.entriesByRepoID, m.Repo.ID)
	delete(s.entriesByRepoName, m.Repo.Name)
}

func (s *AskpassServer) handleAskpassRequest(
	ctx context.Context,
	m askpass.CredentialRequest,
) error {
	repoName, ok := s.parseRepoURL(m.RepoURL)
	if !ok {
		return nil
	}

	entry, ok := s.entriesByRepoName[repoName]
	if !ok {
		return nil
	}

	var value string
	switch m.Credential {
	case askpass.Username:
		value = "x-access-token"
	case askpass.Password:
		token, err := entry.TokenSource.Token()
		if err != nil {
			return err
		}
		value = token.AccessToken
	default:
		return fmt.Errorf("unsupported credential type: %s", m.Credential)
	}

	return minibus.Send(
		ctx,
		askpass.CredentialResponse{
			RequestID:  m.RequestID,
			Repo:       marshalRepo(entry.Repo),
			RepoURL:    m.RepoURL,
			Credential: m.Credential,
			Value:      value,
		},
	)
}

func (s *AskpassServer) parseRepoURL(u *url.URL) (string, bool) {
	if s.Client.BaseURL == nil {
		if u.Host != "github.com" {
			return "", false
		}
	} else if u.Host != s.Client.BaseURL.Host {
		return "", false
	}

	matches := regexp.
		MustCompile(`^/([^/]+/[^/.]+)`).
		FindStringSubmatch(u.Path)

	if matches == nil {
		return "", false
	}

	return matches[1], true
}
