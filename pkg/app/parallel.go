package app

import (
	"context"
	"fmt"
	"sync"
)

// DefaultParallel is the default number of concurrent GitLab API calls.
const DefaultParallel = 5

// parallelMap calls fn for every item with at most limit calls in flight and
// returns the results in the same order as items.
//
// On the first error, the context passed to the remaining calls is canceled,
// no new call is started, and that error is returned.
func parallelMap[T, R any](
	ctx context.Context,
	limit int,
	items []T,
	fn func(context.Context, T) (R, error),
) ([]R, error) {
	if limit < 1 {
		limit = 1
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		results  = make([]R, len(items))
		sem      = make(chan struct{}, limit)
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)

	for i, item := range items {
		select {
		case sem <- struct{}{}:
		case <-runCtx.Done():
		}
		if runCtx.Err() != nil {
			break
		}
		wg.Go(func() {
			defer func() { <-sem }()
			res, err := fn(runCtx, item)
			if err != nil {
				once.Do(func() {
					firstErr = err
					cancel()
				})
				return
			}
			results[i] = res
		})
	}
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("parallel execution interrupted: %w", err)
	}
	return results, nil
}

// flatten concatenates a slice of slices, keeping their order.
func flatten[T any](parts [][]T) []T {
	n := 0
	for _, p := range parts {
		n += len(p)
	}
	out := make([]T, 0, n)
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}
