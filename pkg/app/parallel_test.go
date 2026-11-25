package app

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParallelMap_PreservesOrder(t *testing.T) {
	items := []int{5, 4, 3, 2, 1, 0}

	res, err := parallelMap(context.Background(), 3, items, func(_ context.Context, n int) (int, error) {
		// Later items finish first to make sure the order does not depend on completion.
		time.Sleep(time.Duration(n) * time.Millisecond)
		return n * 10, nil
	})

	require.NoError(t, err)
	assert.Equal(t, []int{50, 40, 30, 20, 10, 0}, res)
}

func TestParallelMap_RespectsLimit(t *testing.T) {
	const limit = 3
	items := make([]int, 20)
	var inFlight, maxInFlight atomic.Int32

	_, err := parallelMap(context.Background(), limit, items, func(_ context.Context, _ int) (struct{}, error) {
		cur := inFlight.Add(1)
		for {
			prev := maxInFlight.Load()
			if cur <= prev || maxInFlight.CompareAndSwap(prev, cur) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		inFlight.Add(-1)
		return struct{}{}, nil
	})

	require.NoError(t, err)
	assert.LessOrEqual(t, maxInFlight.Load(), int32(limit))
	assert.Greater(t, maxInFlight.Load(), int32(1), "calls should run concurrently")
}

func TestParallelMap_StopsOnFirstError(t *testing.T) {
	errBoom := errors.New("boom")
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}
	var calls atomic.Int32

	res, err := parallelMap(context.Background(), 2, items, func(ctx context.Context, n int) (int, error) {
		calls.Add(1)
		if n == 3 {
			return 0, errBoom
		}
		select {
		case <-ctx.Done():
		case <-time.After(10 * time.Millisecond):
		}
		return n, nil
	})

	require.ErrorIs(t, err, errBoom)
	assert.Nil(t, res)
	assert.Less(t, calls.Load(), int32(len(items)), "no new call should start after an error")
}

func TestParallelMap_ParentContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := parallelMap(ctx, 2, []int{1, 2, 3}, func(_ context.Context, n int) (int, error) {
		return n, nil
	})

	require.ErrorIs(t, err, context.Canceled)
}

func TestParallelMap_EmptyAndInvalidLimit(t *testing.T) {
	res, err := parallelMap(context.Background(), 0, []int{1, 2}, func(_ context.Context, n int) (int, error) {
		return n + 1, nil
	})
	require.NoError(t, err)
	assert.Equal(t, []int{2, 3}, res)

	res, err = parallelMap(context.Background(), 4, nil, func(_ context.Context, n int) (int, error) {
		return n, nil
	})
	require.NoError(t, err)
	assert.Empty(t, res)
}

func TestFlatten(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, flatten([][]int{{1}, nil, {2, 3}}))
	assert.Empty(t, flatten[int](nil))
}
