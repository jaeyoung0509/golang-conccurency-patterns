package actor

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestInventoryActorSerializesConcurrentReservations(t *testing.T) {
	actor, err := StartInventoryActor("sku-flash-sale", 10)
	if err != nil {
		t.Fatalf("StartInventoryActor returned error: %v", err)
	}
	defer actor.Stop()

	var successes atomic.Int32
	var wg sync.WaitGroup

	for index := range 25 {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			_, err := actor.Reserve(context.Background(), "order-"+string(rune('a'+index)), 1)
			if err == nil {
				successes.Add(1)
			}
		}(index)
	}

	wg.Wait()

	snapshot, err := actor.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}

	if successes.Load() != 10 {
		t.Fatalf("unexpected success count: got %d want 10", successes.Load())
	}

	if snapshot.Reserved != 10 || snapshot.Available != 0 {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
}

func TestInventoryActorReleaseAndRestock(t *testing.T) {
	actor, err := StartInventoryActor("sku-42", 5)
	if err != nil {
		t.Fatalf("StartInventoryActor returned error: %v", err)
	}
	defer actor.Stop()

	if _, err := actor.Reserve(context.Background(), "order-1", 3); err != nil {
		t.Fatalf("Reserve returned error: %v", err)
	}

	if _, err := actor.Release(context.Background(), "order-1", 1); err != nil {
		t.Fatalf("Release returned error: %v", err)
	}

	snapshot, err := actor.Restock(context.Background(), 4)
	if err != nil {
		t.Fatalf("Restock returned error: %v", err)
	}

	if snapshot.OnHand != 9 || snapshot.Reserved != 2 || snapshot.Available != 7 {
		t.Fatalf("unexpected snapshot after release/restock: %#v", snapshot)
	}
}

func TestInventoryActorRejectsCommandsAfterStop(t *testing.T) {
	actor, err := StartInventoryActor("sku-42", 5)
	if err != nil {
		t.Fatalf("StartInventoryActor returned error: %v", err)
	}

	actor.Stop()

	_, err = actor.Snapshot(context.Background())
	if !errors.Is(err, ErrActorStopped) {
		t.Fatalf("expected ErrActorStopped, got %v", err)
	}
}
