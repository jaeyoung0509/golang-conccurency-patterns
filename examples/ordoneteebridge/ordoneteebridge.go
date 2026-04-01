package ordoneteebridge

import "context"

type FeedItem struct {
	TenantID string
	Offset   int
	Kind     string
}

// OrDone forwards values until the input closes or the parent context is done.
// It is the smallest useful wrapper for cancellation-safe channel reads.
func OrDone[T any](ctx context.Context, in <-chan T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case value, ok := <-in:
				if !ok {
					return
				}

				select {
				case out <- value:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

// Tee duplicates every input value to both outputs while still honoring
// cancellation. Each loop nils the channel that already received the value so
// one slow consumer does not get the value twice.
func Tee[T any](ctx context.Context, in <-chan T) (<-chan T, <-chan T) {
	left := make(chan T)
	right := make(chan T)

	go func() {
		defer close(left)
		defer close(right)

		for value := range OrDone(ctx, in) {
			leftOut := left
			rightOut := right

			for range 2 {
				select {
				case <-ctx.Done():
					return
				case leftOut <- value:
					leftOut = nil
				case rightOut <- value:
					rightOut = nil
				}
			}
		}
	}()

	return left, right
}

// Bridge flattens a channel of channels into one stream. It drains each inner
// stream through OrDone so outer cancellation tears the whole composition down.
func Bridge[T any](ctx context.Context, streams <-chan <-chan T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for {
			var stream <-chan T

			select {
			case <-ctx.Done():
				return
			case next, ok := <-streams:
				if !ok {
					return
				}
				stream = next
			}

			for value := range OrDone(ctx, stream) {
				select {
				case out <- value:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

// MergeTenantFeeds is a concrete bridge example: each tenant produces its own
// feed, and the caller wants one flattened stream to consume.
func MergeTenantFeeds(ctx context.Context, feeds ...[]FeedItem) <-chan FeedItem {
	streams := make(chan (<-chan FeedItem))

	go func() {
		defer close(streams)

		for _, feed := range feeds {
			feed := feed
			stream := make(chan FeedItem, len(feed))

			for _, item := range feed {
				stream <- item
			}
			close(stream)

			select {
			case streams <- stream:
			case <-ctx.Done():
				return
			}
		}
	}()

	return Bridge(ctx, streams)
}
