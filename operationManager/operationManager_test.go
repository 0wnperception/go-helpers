package operationManager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkOperationManager(b *testing.B) {
	type Args struct {
		num int
	}
	p := &Args{}
	m := NewOperationManager[string, *Args](b.N)
	b.Run("3 op", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.AddOperation("+1", func(ctx context.Context, args *Args) error {
				args.num += 1
				return nil
			})
		}
		m.Run(context.Background(), p)
	})
}

func TestOperationManager(t *testing.T) {
	r := require.New(t)

	type Args struct {
		num int
	}
	t.Run("1 operation", func(t *testing.T) {
		m := NewOperationManager[string, *Args](1)
		p := &Args{}
		m.AddOperation("+1", func(ctx context.Context, args *Args) error {
			args.num++
			return nil
		})
		t.Log("op added")
		r.Len(m.operations, 1)
		r.Equal(m.operationsQueue.Len(), 1)
		t.Log("operations array len is good")
		m.Run(context.Background(), p)
		r.Equal(1, m.operationsQueue.Len())
		r.Equal(1, p.num)
	})

	t.Run("3 operations", func(t *testing.T) {
		m := NewOperationManager[string, *Args](3)
		p := &Args{}
		m.AddOperation("+5", func(ctx context.Context, args *Args) error {
			args.num += 5
			return nil
		})
		m.AddOperation("-2", func(ctx context.Context, args *Args) error {
			args.num -= 2
			return nil
		})
		m.AddOperation("+4", func(ctx context.Context, args *Args) error {
			args.num += 4
			return nil
		})
		t.Log("op added")
		r.Len(m.operations, 3)
		t.Log("operations array len is good")
		m.Run(context.Background(), p)
		r.Equal(p.num, 7)
		r.Len(m.operations, 3)
		r.Equal(m.operationsQueue.Len(), 3)
	})
}
