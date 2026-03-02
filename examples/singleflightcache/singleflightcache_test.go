package singleflightcache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPriceServiceDeduplicatesConcurrentMisses(t *testing.T) {
	var calls atomic.Int32

	service, err := NewPriceService(func(_ context.Context, sku string) (PriceQuote, error) {
		calls.Add(1)
		time.Sleep(20 * time.Millisecond)
		return PriceQuote{SKU: sku, Currency: "USD", UnitCents: 1999, Source: "pricing-db"}, nil
	})
	if err != nil {
		t.Fatalf("NewPriceService returned error: %v", err)
	}

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			quote, err := service.Get(context.Background(), "sku-42")
			if err != nil {
				t.Errorf("Get returned error: %v", err)
				return
			}

			if quote.SKU != "sku-42" {
				t.Errorf("unexpected quote: %#v", quote)
			}
		}()
	}

	wg.Wait()

	if calls.Load() != 1 {
		t.Fatalf("expected one loader call, got %d", calls.Load())
	}
}

func TestPriceServiceFollowerCanCancelWhileLeaderContinues(t *testing.T) {
	var calls atomic.Int32
	leaderStarted := make(chan struct{})
	leaderMayFinish := make(chan struct{})

	service, err := NewPriceService(func(ctx context.Context, sku string) (PriceQuote, error) {
		calls.Add(1)
		close(leaderStarted)

		select {
		case <-leaderMayFinish:
		case <-ctx.Done():
			return PriceQuote{}, ctx.Err()
		}

		return PriceQuote{SKU: sku, Currency: "USD", UnitCents: 2400, Source: "pricing-db"}, nil
	})
	if err != nil {
		t.Fatalf("NewPriceService returned error: %v", err)
	}

	leaderDone := make(chan error, 1)
	go func() {
		_, err := service.Get(context.Background(), "sku-42")
		leaderDone <- err
	}()

	<-leaderStarted

	followerCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err = service.Get(followerCtx, "sku-42")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded for follower, got %v", err)
	}

	close(leaderMayFinish)

	select {
	case err := <-leaderDone:
		if err != nil {
			t.Fatalf("leader returned error: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("leader did not finish")
	}

	if calls.Load() != 1 {
		t.Fatalf("expected one loader call, got %d", calls.Load())
	}
}

func TestPriceServiceDoesNotCacheErrors(t *testing.T) {
	var calls atomic.Int32
	var fail atomic.Bool
	fail.Store(true)

	service, err := NewPriceService(func(_ context.Context, sku string) (PriceQuote, error) {
		calls.Add(1)

		if fail.Load() {
			return PriceQuote{}, errors.New("pricing database unavailable")
		}

		return PriceQuote{SKU: sku, Currency: "USD", UnitCents: 3100, Source: "pricing-db"}, nil
	})
	if err != nil {
		t.Fatalf("NewPriceService returned error: %v", err)
	}

	if _, err := service.Get(context.Background(), "sku-42"); err == nil {
		t.Fatal("expected first call to fail")
	}

	fail.Store(false)

	quote, err := service.Get(context.Background(), "sku-42")
	if err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}

	if quote.UnitCents != 3100 {
		t.Fatalf("unexpected quote: %#v", quote)
	}

	if calls.Load() != 2 {
		t.Fatalf("expected two loader calls, got %d", calls.Load())
	}
}
