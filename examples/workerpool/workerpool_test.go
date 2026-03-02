package workerpool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestGenerateQuotesPreservesInputOrder(t *testing.T) {
	orders := []Order{
		{ID: "ord-101", DistanceKM: 550, WeightKG: 12.5, Priority: "priority"},
		{ID: "ord-102", DistanceKM: 120, WeightKG: 2.1, Priority: "standard"},
		{ID: "ord-103", DistanceKM: 960, WeightKG: 8.4, Priority: "priority"},
	}

	quotes, err := GenerateQuotes(context.Background(), orders, 3, func(_ context.Context, order Order) (ShipmentQuote, error) {
		time.Sleep(time.Duration(4-order.DistanceKM/300) * 5 * time.Millisecond)

		return ShipmentQuote{
			OrderID:    order.ID,
			Carrier:    "fastship",
			PriceCents: order.DistanceKM*7 + int(order.WeightKG*100),
			SLAHours:   24,
		}, nil
	})
	if err != nil {
		t.Fatalf("GenerateQuotes returned error: %v", err)
	}

	for index, quote := range quotes {
		if quote.OrderID != orders[index].ID {
			t.Fatalf("quote %d order mismatch: got %s want %s", index, quote.OrderID, orders[index].ID)
		}
	}
}

func TestGenerateQuotesRespectsWorkerLimit(t *testing.T) {
	orders := make([]Order, 18)
	for index := range orders {
		orders[index] = Order{ID: "ord", DistanceKM: 80 + index, WeightKG: 1.5}
	}

	var inFlight atomic.Int32
	var maxInFlight atomic.Int32

	_, err := GenerateQuotes(context.Background(), orders, 4, func(_ context.Context, order Order) (ShipmentQuote, error) {
		current := inFlight.Add(1)
		defer inFlight.Add(-1)

		for {
			seen := maxInFlight.Load()
			if current <= seen || maxInFlight.CompareAndSwap(seen, current) {
				break
			}
		}

		time.Sleep(12 * time.Millisecond)

		return ShipmentQuote{OrderID: order.ID, Carrier: "fastship"}, nil
	})
	if err != nil {
		t.Fatalf("GenerateQuotes returned error: %v", err)
	}

	if maxInFlight.Load() > 4 {
		t.Fatalf("worker limit exceeded: got %d", maxInFlight.Load())
	}
}

func TestGenerateQuotesCancelsSlowJobsAfterError(t *testing.T) {
	cancelObserved := make(chan struct{}, 1)

	orders := []Order{
		{ID: "slow-a"},
		{ID: "bad"},
		{ID: "slow-b"},
	}

	_, err := GenerateQuotes(context.Background(), orders, 2, func(ctx context.Context, order Order) (ShipmentQuote, error) {
		if order.ID == "bad" {
			return ShipmentQuote{}, errors.New("carrier API unavailable")
		}

		<-ctx.Done()
		select {
		case cancelObserved <- struct{}{}:
		default:
		}

		return ShipmentQuote{}, ctx.Err()
	})
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	select {
	case <-cancelObserved:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("slow job did not observe cancellation")
	}
}
