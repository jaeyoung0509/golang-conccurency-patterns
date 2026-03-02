package gracefulshutdown

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestShutdownDrainsAcceptedEvents(t *testing.T) {
	var (
		mu      sync.Mutex
		handled []string
	)

	processor, err := NewProcessor(2, 4, func(event Event) error {
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		handled = append(handled, event.OrderID)
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("NewProcessor returned error: %v", err)
	}

	events := []Event{
		{OrderID: "ord-1", Kind: "placed"},
		{OrderID: "ord-2", Kind: "paid"},
		{OrderID: "ord-3", Kind: "packed"},
	}

	for _, event := range events {
		if err := processor.Submit(context.Background(), event); err != nil {
			t.Fatalf("Submit returned error: %v", err)
		}
	}

	if err := processor.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	mu.Lock()
	sorted := append([]string(nil), handled...)
	mu.Unlock()
	slices.Sort(sorted)

	if len(sorted) != len(events) {
		t.Fatalf("handled %d events, want %d", len(sorted), len(events))
	}
}

func TestSubmitReturnsErrClosedAfterShutdown(t *testing.T) {
	processor, err := NewProcessor(1, 1, func(Event) error { return nil })
	if err != nil {
		t.Fatalf("NewProcessor returned error: %v", err)
	}

	if err := processor.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	err = processor.Submit(context.Background(), Event{OrderID: "ord-1"})
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("Submit error = %v, want ErrClosed", err)
	}
}

func TestShutdownHonorsDeadline(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	processor, err := NewProcessor(1, 1, func(Event) error {
		close(started)
		<-release
		return nil
	})
	if err != nil {
		t.Fatalf("NewProcessor returned error: %v", err)
	}

	if err := processor.Submit(context.Background(), Event{OrderID: "ord-1"}); err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	<-started

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := processor.Shutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown error = %v, want deadline exceeded", err)
	}

	close(release)

	if err := processor.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown retry returned error: %v", err)
	}
}
