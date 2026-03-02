---
title: Fan-Out / Fan-In
description: Query several backends in parallel and merge the useful responses into a single report.
---

# Fan-Out / Fan-In

Fan-out / fan-in is the right pattern when you want to ask the same question to many independent sources and then combine the answers.

The example in `examples/fanoutfanin` checks product inventory across multiple warehouses at the same time.

Unlike the worker pool example, this pattern does not treat one backend failure as a total failure by default. Partial results are still valuable.

## Example scenario

- fan out: send one lookup to each warehouse,
- fan in: collect all responses in one channel,
- aggregate: sort viable options and record failures separately.

```mermaid
flowchart LR
    A["SKU request"] --> B["Warehouse lookups"]
    B --> C1["London"]
    B --> C2["Berlin"]
    B --> C3["Seoul"]
    C1 --> D["results channel"]
    C2 --> D
    C3 --> D
    D --> E["InventoryReport"]
    E --> F["Best option"]
    E --> G["Failure summary"]
```

## Key implementation idea

The report preserves good data and bad data separately.

```go
func CollectInventory(ctx context.Context, sku string, lookups map[string]WarehouseLookup) (InventoryReport, error) {
    results := make(chan result, len(lookups))

    for name, lookup := range lookups {
        go func(name string, lookup WarehouseLookup) {
            stock, err := lookup(ctx, sku)
            results <- result{name: name, stock: stock, err: err}
        }(name, lookup)
    }

    for item := range results {
        if item.err != nil {
            report.Failures[item.name] = item.err.Error()
            continue
        }

        report.Options = append(report.Options, item.stock)
    }

    sortOptions(report.Options)
    return report, nil
}
```

That contract is practical because warehouse outages are common, but that does not mean the caller should lose all availability information.

## What the tests prove

The tests verify that:

- successful lookups are kept even when one warehouse fails,
- the report chooses the best option deterministically,
- full cancellation still propagates from the caller context.

## Important design choice

This example intentionally does **not** cancel on the first warehouse failure.

That is the difference between a useful aggregation flow and an overly fragile one.
If your product can work with partial inventory visibility, keep partial results.

## Common mistakes

### Treating all downstream errors equally

Some systems need "all or nothing". Others need "best effort". Decide which one you are building before you write the code.

### Returning unsorted merged results

Concurrent completion order is rarely the same as business priority. Sort after fan-in so callers receive stable output.

### Unbounded fan-out

This example fans out once per warehouse. That is usually small and known. If the target set can grow large, combine this pattern with a worker pool or semaphore.
