package channels

import (
	"context"
	"sync"
)

func ContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func ContextNotDone(ctx context.Context) bool { return !ContextDone(ctx) }

func AddToChannel[T any](c chan<- T, t T) error {
	return AddToChannelWithContext(context.Background(), c, t)
}

func CopyChannel[T any](c <-chan T) (<-chan T, <-chan T) {
	return CopyChannelWithContext(context.Background(), c)
}

func SliceToChannel[T any](buffersize int, slice []T) <-chan T {
	return SliceToChannelWithContext(context.Background(), buffersize, slice)
}

func ChannelToSlice[T any](c <-chan T) []T {
	return ChannelToSliceWithContext(context.Background(), c)
}

func Multiplex[T1, T2 any](c <-chan T1, numworkers int, f func(T1) T2) <-chan T2 {
	return MultiplexWithContext(context.Background(), c, numworkers, f)
}

func SliceToChannelWithContext[T any](ctx context.Context, buffersize int, slice []T) <-chan T {
	results := make(chan T, buffersize)
	go func() {
		defer close(results)
		for _, t := range slice {
			if err := AddToChannelWithContext(ctx, results, t); err != nil {
				return
			}
		}
	}()
	return results
}

func ChannelToSliceWithContext[T any](ctx context.Context, c <-chan T) []T {
	results := make([]T, 0)
loop:
	for ContextNotDone(ctx) {
		select {
		case <-ctx.Done():
			break loop
		case t, ok := <-c:
			if !ok {
				break loop
			} else {
				results = append(results, t)
			}
		}
	}
	return results
}

func AddToChannelWithContext[T any](ctx context.Context, c chan<- T, t T) error {
	select {
	case <-ctx.Done():
		return context.DeadlineExceeded
	case c <- t:
		return nil
	}
}

func CopyChannelWithContext[T any](ctx context.Context, c <-chan T) (<-chan T, <-chan T) {
	a, b := make(chan T, cap(c)), make(chan T, cap(c))
	go func() {
		defer close(a)
		defer close(b)
		for ContextNotDone(ctx) {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-c:
				if !ok {
					return
				} else if err := AddToChannelWithContext(ctx, a, t); err != nil {
					return
				} else if err := AddToChannelWithContext(ctx, b, t); err != nil {
					return
				}
			}
		}
	}()
	return a, b
}

func MultiplexWithContext[T1, T2 any](ctx context.Context, c <-chan T1, numworkers int, f func(T1) T2) <-chan T2 {
	results := make(chan T2, cap(c))
	go func() {
		defer close(results)
		var wg sync.WaitGroup
		for i := 0; i < numworkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for ContextNotDone(ctx) {
					select {
					case <-ctx.Done():
						return
					case t, ok := <-c:
						if !ok {
							return
						} else if err := AddToChannelWithContext(ctx, results, f(t)); err != nil {
							return
						}
					}
				}
			}()
		}
		wg.Wait()
	}()
	return results
}
