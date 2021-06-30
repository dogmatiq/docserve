package analyzer

import (
	"context"
	"sync"

	"github.com/google/go-github/v35/github"
)

type Orchestrator struct {
	Analyzer *Analyzer
	Remover  *Remover

	once  sync.Once
	queue chan queueItem
}

type queueItem struct {
	repo   *github.Repository
	remove bool
}

func (o *Orchestrator) Run(ctx context.Context) error {
	o.init()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item := <-o.queue:
			if err := o.handle(ctx, item); err != nil {
				return err
			}
		}
	}
}

func (o *Orchestrator) EnqueueAnalyis(ctx context.Context, r *github.Repository) error {
	return o.enqueue(ctx, r, false)
}

func (o *Orchestrator) EnqueueRemoval(ctx context.Context, r *github.Repository) error {
	return o.enqueue(ctx, r, true)
}

func (o *Orchestrator) handle(ctx context.Context, item queueItem) error {
	if item.remove {
		return o.Remover.Remove(ctx, item.repo)
	}

	return o.Analyzer.Analyze(ctx, item.repo)
}

func (o *Orchestrator) enqueue(ctx context.Context, r *github.Repository, remove bool) error {
	o.init()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case o.queue <- queueItem{r, remove}:
		return nil
	}
}

func (o *Orchestrator) init() {
	o.once.Do(func() {
		o.queue = make(chan queueItem, 100)
	})
}
