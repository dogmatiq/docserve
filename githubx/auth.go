package githubx

import (
	"context"
	"crypto/rsa"
	"strconv"
	"time"

	"github.com/dogmatiq/linger"
	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

// AppTokenSource is an implementation of oauth2.TokenSource that generates
// GitHub API tokens that authenticate a GitHub application.
type AppTokenSource struct {
	AppID      int64
	PrivateKey *rsa.PrivateKey
	TTL        time.Duration
}

func (s *AppTokenSource) Token() (*oauth2.Token, error) {
	ttl := linger.MustCoalesce(s.TTL, 1*time.Minute)
	expiresAt := time.Now().Add(ttl)

	header := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
	}

	claims := &jws.ClaimSet{
		Iss: strconv.FormatInt(s.AppID, 10),
		Exp: expiresAt.Unix(),
	}

	// TODO: replace use of deprecated jws package with something else.
	token, err := jws.Encode(header, claims, s.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: token,
		Expiry:      expiresAt,
	}, nil
}

// InstallationTokenSource is an installation of oauth2.TokenSource that
// generates GitHub API tokens that authenticate a specific "installation" of a
// GitHub application.
type InstallationTokenSource struct {
	AppClient      *github.Client
	InstallationID int64
	Options        *github.InstallationTokenOptions
	Timeout        time.Duration
}

func (s *InstallationTokenSource) Token() (*oauth2.Token, error) {
	ctx, cancel := linger.ContextWithTimeout(context.Background(), s.Timeout, 10*time.Second)
	defer cancel()

	token, _, err := s.AppClient.Apps.CreateInstallationToken(
		ctx,
		s.InstallationID,
		s.Options,
	)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: token.GetToken(),
		Expiry:      token.GetExpiresAt(),
	}, nil
}
