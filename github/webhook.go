package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dogmatiq/dapper"
	"github.com/google/go-github/v35/github"
)

type HookHandler struct {
	Secret []byte
}

func (h *HookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, h.Secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hookType := github.WebHookType(r)
	event, err := github.ParseWebHook(hookType, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dapper.Print(event)

	if err := h.handleEvent(r.Context(), event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HookHandler) handleEvent(ctx context.Context, event interface{}) error {
	// TODO: handle repository rename
	switch event := event.(type) {
	case *github.InstallationEvent:
		return h.handleInstallationEvent(ctx, event)
	case *github.InstallationRepositoriesEvent:
		return h.handleInstallationRepositoriesEvent(ctx, event)
	case *github.PushEvent:
		return h.handlePushEvent(ctx, event)
	}

	return nil
}

func (h *HookHandler) handleInstallationEvent(ctx context.Context, event *github.InstallationEvent) error {
	if event.Action == nil {
		return nil
	}

	switch *event.Action {
	case "created", "unsuspend", "new_permissions_accepted":
		for _, r := range event.Repositories {
			if err := h.analyzeRepository(ctx, r.GetFullName()); err != nil {
				return err
			}
		}
	case "deleted", "suspend":
		for _, r := range event.Repositories {
			if err := h.removeRepository(ctx, r.GetFullName()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *HookHandler) handleInstallationRepositoriesEvent(ctx context.Context, event *github.InstallationRepositoriesEvent) error {
	for _, r := range event.RepositoriesRemoved {
		if err := h.removeRepository(ctx, r.GetFullName()); err != nil {
			return err
		}
	}

	for _, r := range event.RepositoriesAdded {
		if err := h.analyzeRepository(ctx, r.GetFullName()); err != nil {
			return err
		}
	}

	return nil
}

func (h *HookHandler) handlePushEvent(ctx context.Context, event *github.PushEvent) error {
	repo := event.GetRepo()
	defaultRef := "refs/heads/" + repo.GetDefaultBranch()

	if event.GetRef() != defaultRef {
		return nil
	}

	return h.analyzeRepository(ctx, repo.GetFullName())
}

func (h *HookHandler) analyzeRepository(ctx context.Context, slug string) error {
	fmt.Println("analyze", slug)
	return nil
}

func (h *HookHandler) removeRepository(ctx context.Context, slug string) error {
	fmt.Println("remove", slug)
	return nil
}
