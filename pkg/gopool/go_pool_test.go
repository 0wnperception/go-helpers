package gopool

import (
	"context"
	"errors"
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

		err = p.Wait()
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		p, err := New()
		require.NoError(t, err)

		tmpErr := errors.New("error")

		err = p.Run(ctx, func(ctx context.Context) error {
			time.Sleep(time.Second)

			return nil
		})
		require.NoError(t, err)

		err = p.Run(ctx, func(ctx context.Context) error {
			return tmpErr
		})
		require.NoError(t, err)

		err = p.Wait()
		require.ErrorIs(t, err, tmpErr)
	})
}
