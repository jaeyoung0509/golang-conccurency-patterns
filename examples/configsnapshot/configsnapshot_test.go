package configsnapshot

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBootstrapRunsLoaderOnlyOnce(t *testing.T) {
	var calls atomic.Int32

	store, err := NewStore(func(_ context.Context) (Snapshot, error) {
		calls.Add(1)
		time.Sleep(10 * time.Millisecond)

		return Snapshot{
			Version: 1,
			FeatureFlags: map[string]bool{
				"beta-search": true,
			},
			RegionBudget: map[string]int{
				"eu-west-1": 24,
			},
		}, nil
	})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 8)

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- store.Bootstrap(context.Background())
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("Bootstrap returned error: %v", err)
		}
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("loader calls = %d, want 1", got)
	}

	if !store.Enabled("beta-search") {
		t.Fatalf("expected published snapshot to enable beta-search")
	}
}

func TestReloadFailureKeepsLastGoodSnapshot(t *testing.T) {
	var calls atomic.Int32

	store, err := NewStore(func(_ context.Context) (Snapshot, error) {
		switch calls.Add(1) {
		case 1:
			return Snapshot{
				Version: 1,
				FeatureFlags: map[string]bool{
					"beta-search": true,
				},
				RegionBudget: map[string]int{
					"us-east-1": 10,
				},
			}, nil
		default:
			return Snapshot{}, errors.New("config backend unavailable")
		}
	})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	if err := store.Bootstrap(context.Background()); err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	if err := store.Reload(context.Background()); err == nil {
		t.Fatalf("Reload error = nil, want backend error")
	}

	snapshot, ok := store.Current()
	if !ok {
		t.Fatalf("Current returned no snapshot after successful bootstrap")
	}
	if snapshot.Version != 1 {
		t.Fatalf("snapshot version = %d, want 1", snapshot.Version)
	}
	if !snapshot.FeatureFlags["beta-search"] {
		t.Fatalf("expected previous snapshot to remain active after reload failure")
	}
}

func TestCurrentReturnsIsolatedCopy(t *testing.T) {
	store, err := NewStore(func(_ context.Context) (Snapshot, error) {
		return Snapshot{
			Version: 2,
			FeatureFlags: map[string]bool{
				"beta-search": true,
			},
			RegionBudget: map[string]int{
				"ap-south-1": 15,
			},
		}, nil
	})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	if err := store.Bootstrap(context.Background()); err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	snapshot, ok := store.Current()
	if !ok {
		t.Fatalf("Current returned no snapshot")
	}

	snapshot.FeatureFlags["beta-search"] = false
	snapshot.RegionBudget["ap-south-1"] = 1

	if !store.Enabled("beta-search") {
		t.Fatalf("mutating returned snapshot should not alter stored snapshot")
	}

	budget, ok := store.RegionBudgetFor("ap-south-1")
	if !ok {
		t.Fatalf("expected stored region budget to exist")
	}
	if budget != 15 {
		t.Fatalf("stored budget = %d, want 15", budget)
	}
}
