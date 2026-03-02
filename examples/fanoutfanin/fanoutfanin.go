package fanoutfanin

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type WarehouseStock struct {
	Warehouse string
	Available int
	ETAHours  int
	CheckedAt time.Time
}

type WarehouseLookup func(context.Context, string) (WarehouseStock, error)

type InventoryReport struct {
	SKU      string
	Options  []WarehouseStock
	Failures map[string]string
}

func (report InventoryReport) BestOption() (WarehouseStock, bool) {
	if len(report.Options) == 0 {
		return WarehouseStock{}, false
	}

	return report.Options[0], true
}

func CollectInventory(ctx context.Context, sku string, lookups map[string]WarehouseLookup) (InventoryReport, error) {
	if sku == "" {
		return InventoryReport{}, errors.New("sku must not be empty")
	}

	if len(lookups) == 0 {
		return InventoryReport{}, errors.New("lookups must not be empty")
	}

	type result struct {
		name  string
		stock WarehouseStock
		err   error
	}

	results := make(chan result, len(lookups))

	var workersWG sync.WaitGroup
	for name, lookup := range lookups {
		workersWG.Add(1)

		go func(name string, lookup WarehouseLookup) {
			defer workersWG.Done()

			stock, err := lookup(ctx, sku)
			if err == nil && stock.Warehouse == "" {
				stock.Warehouse = name
			}

			select {
			case results <- result{name: name, stock: stock, err: err}:
			case <-ctx.Done():
			}
		}(name, lookup)
	}

	go func() {
		workersWG.Wait()
		close(results)
	}()

	report := InventoryReport{
		SKU:      sku,
		Failures: make(map[string]string),
	}

	for item := range results {
		if item.err != nil {
			report.Failures[item.name] = item.err.Error()
			continue
		}

		report.Options = append(report.Options, item.stock)
	}

	sort.Slice(report.Options, func(i, j int) bool {
		left := report.Options[i]
		right := report.Options[j]

		if (left.Available > 0) != (right.Available > 0) {
			return left.Available > 0
		}

		if left.ETAHours != right.ETAHours {
			return left.ETAHours < right.ETAHours
		}

		if left.Available != right.Available {
			return left.Available > right.Available
		}

		return left.Warehouse < right.Warehouse
	})

	if len(report.Options) > 0 {
		return report, nil
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	return report, errors.New("all inventory lookups failed")
}
