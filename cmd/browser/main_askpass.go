package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dogmatiq/browser/components/askpass"
	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
)

func runAskpass(ctx context.Context) error {
	id := uuid.New()

	if len(os.Args) < 2 {
		return fmt.Errorf("[request %d] expected at least one argument", id)
	}

	var field askpass.Field
	switch {
	case strings.HasPrefix(os.Args[1], "Username "):
		field = askpass.Username
	case strings.HasPrefix(os.Args[1], "Password "):
		field = askpass.Password
	default:
		return fmt.Errorf("[request %s] unexpected prompt: %s", id, os.Args[1])
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("[request %s] unable to open repository: %w", id, err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("[request %s] unable to find 'origin' remote: %w", id, err)
	}

	value, err := askpass.Ask(
		ctx,
		os.Getenv("ASKPASS_ADDR"),
		id,
		remote.Config().URLs[0],
		field,
	)
	if err != nil {
		return fmt.Errorf("[request %s] %w", id, err)
	}

	fmt.Println(value)

	return nil
}
