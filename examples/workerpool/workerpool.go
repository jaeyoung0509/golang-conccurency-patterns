package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Order struct {
	ID         string
	DistanceKM int
	WeightKG   float64
	Priority   string
}

type ShipmentQuote struct {
	OrderID    string
	Carrier    string
	PriceCents int
	SLAHours   int
}

type QuoteFunc func(context.Context, Order) (ShipmentQuote, error)

func GenerateQuotes(ctx context.Context, orders []Order, workers int, quoteFn QuoteFunc) ([]ShipmentQuote, error) {
	if workers < 1 {
		return nil, errors.New("workers must be at least 1")
	}

	if quoteFn == nil {
		return nil, errors.New("quoteFn must not be nil")
	}

	if len(orders) == 0 {
		return []ShipmentQuote{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type job struct {
		index int
		order Order
	}

	type result struct {
		index int
		quote ShipmentQuote
		err   error
	}

	jobs := make(chan job)
	// Buffering by worker count lets finished workers report back without
	// immediately blocking on a briefly slow collector.
	results := make(chan result, workers)

	var workersWG sync.WaitGroup
	worker := func() {
		defer workersWG.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-jobs:
				if !ok {
					return
				}

				quote, err := quoteFn(ctx, item.order)

				// Every result carries the original index so the collector can
				// reconstruct stable output order even if workers finish out of order.
				select {
				case results <- result{index: item.index, quote: quote, err: err}:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	workersWG.Add(workers)
	for range workers {
		go worker()
	}

	go func() {
		defer close(jobs)

		for index, order := range orders {
			select {
			case jobs <- job{index: index, order: order}:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		workersWG.Wait()
		close(results)
	}()

	quotes := make([]ShipmentQuote, len(orders))
	var firstErr error
	completed := 0

	for item := range results {
		completed++

		if item.err != nil && firstErr == nil {
			// The collector owns failure policy: first hard error cancels the whole batch.
			firstErr = fmt.Errorf("quote order %s: %w", orders[item.index].ID, item.err)
			cancel()
			continue
		}

		if firstErr == nil {
			quotes[item.index] = item.quote
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	if completed != len(orders) {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		return nil, fmt.Errorf("collected %d quotes for %d orders", completed, len(orders))
	}

	return quotes, nil
}
