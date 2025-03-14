package gopool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGoPool(t *testing.T) {
	ctx := context.Background()

	t.Run("no error", func(t *testing.T) {
		p, err := New()
		require.NoError(t, err)

		err = p.Run(ctx, func(ctx context.Context) error {
			time.Sleep(time.Second)

			return nil
		})
		require.NoError(t, err)

		err = p.WaitFinish()
		require.NoError(t, err)
	})

	t.Run("run error started", func(t *testing.T) {
		p, err := New()
		require.NoError(t, err)

		err = p.Run(ctx, func(ctx context.Context) error {
			time.Sleep(time.Second)

			return nil
		})
		require.NoError(t, err)

		err = p.Run(ctx, func(ctx context.Context) error {
			return nil
		})
		require.ErrorIs(t, err, ErrAlreadyStarted)

		err = p.WaitFinish()
		require.NoError(t, err)
	})

	t.Run("run error finished", func(t *testing.T) {
		p, err := New()
		require.NoError(t, err)

		err = p.Run(ctx, func(ctx context.Context) error {
			return nil
		})
		require.NoError(t, err)

		err = p.WaitFinish()
		require.NoError(t, err)

		err = p.Run(ctx, func(ctx context.Context) error {
			return nil
		})
		require.ErrorIs(t, err, ErrAlreadyFinished)
	})

	t.Run("finish error finished", func(t *testing.T) {
		p, err := New()
		require.NoError(t, err)

		err = p.Run(ctx, func(ctx context.Context) error {
			return nil
		})
		require.NoError(t, err)

		err = p.WaitFinish()
		require.NoError(t, err)

		err = p.WaitFinish()
		require.ErrorIs(t, err, ErrAlreadyFinished)
	})

	t.Run("cancel", func(t *testing.T) {
		p, err := New()
		require.NoError(t, err)

		err = p.Run(ctx, func(ctx context.Context) error {
			t := time.NewTimer(30 * time.Second)

			for {
				select {
				case <-t.C:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
		require.NoError(t, err)

		p.Cancel()

		err = p.WaitFinish()
		require.ErrorIs(t, err, context.Canceled)
	})
}
