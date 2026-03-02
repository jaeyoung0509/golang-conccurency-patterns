package fanoutfanin

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCollectInventoryReturnsBestOptionAndFailures(t *testing.T) {
	now := time.Date(2026, 3, 2, 9, 30, 0, 0, time.UTC)

	report, err := CollectInventory(context.Background(), "sku-42", map[string]WarehouseLookup{
		"london": func(context.Context, string) (WarehouseStock, error) {
			return WarehouseStock{Warehouse: "london", Available: 12, ETAHours: 24, CheckedAt: now}, nil
		},
		"berlin": func(context.Context, string) (WarehouseStock, error) {
			return WarehouseStock{}, errors.New("warehouse API throttled")
		},
		"seoul": func(context.Context, string) (WarehouseStock, error) {
			return WarehouseStock{Warehouse: "seoul", Available: 4, ETAHours: 6, CheckedAt: now}, nil
		},
	})
	if err != nil {
		t.Fatalf("CollectInventory returned error: %v", err)
	}

	if len(report.Options) != 2 {
		t.Fatalf("unexpected option count: got %d want 2", len(report.Options))
	}

	best, ok := report.BestOption()
	if !ok {
		t.Fatal("expected best option but got none")
	}

	if best.Warehouse != "seoul" {
		t.Fatalf("unexpected best option: %#v", best)
	}

	if report.Failures["berlin"] == "" {
		t.Fatalf("expected berlin failure to be recorded: %#v", report.Failures)
	}
}

func TestCollectInventoryFailsWhenAllLookupsFail(t *testing.T) {
	_, err := CollectInventory(context.Background(), "sku-42", map[string]WarehouseLookup{
		"a": func(context.Context, string) (WarehouseStock, error) { return WarehouseStock{}, errors.New("down") },
		"b": func(context.Context, string) (WarehouseStock, error) { return WarehouseStock{}, errors.New("down") },
	})
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

func TestCollectInventoryPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := CollectInventory(ctx, "sku-42", map[string]WarehouseLookup{
		"a": func(ctx context.Context, _ string) (WarehouseStock, error) {
			<-ctx.Done()
			return WarehouseStock{}, ctx.Err()
		},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
