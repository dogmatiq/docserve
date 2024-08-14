package worker

import (
	"context"

	"github.com/dogmatiq/minibus"
	"golang.org/x/sync/errgroup"
)

// RunPool runs a pool of n workers that process messages of type M by calling
// fn.
func RunPool[M any](
	ctx context.Context,
	n int,
	fn func(ctx context.Context, workerID int, message M) error,
) error {
	tasks := make(chan M)
	var queue queue[M]

	minibus.Subscribe[M](ctx)
	minibus.Ready(ctx)

	g, ctx := errgroup.WithContext(ctx)
	for workerID := 1; workerID <= n; workerID++ {
		g.Go(func() error {
			for m := range tasks {
				if err := fn(ctx, workerID, m); err != nil {
					return err
				}
			}
			return nil
		})
	}

	in := minibus.Inbox(ctx)

	for {
		var out chan<- M

		head, ok := queue.Head()
		if ok {
			out = tasks
		}

		select {
		case <-ctx.Done():
			close(tasks)
			return g.Wait()

		case out <- head:
			queue.Pop()

		case tail, ok := <-in:
			if ok {
				queue.Push(tail.(M))
			} else {
				in = nil
			}
		}
	}
}
