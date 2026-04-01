package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type CheckoutEvent struct {
	ID              string
	CustomerID      string
	Region          string
	TotalCents      int
	NewDevice       bool
	PaymentFailures int
}

type RiskSignal struct {
	CheckoutEvent
	Score   int
	Reasons []string
}

type Alert struct {
	CheckoutID string
	Severity   string
	Score      int
	Summary    string
}

type RiskScorer func(context.Context, CheckoutEvent) (RiskSignal, error)

func RunAlertPipeline(ctx context.Context, events []CheckoutEvent, workers int, threshold int, scorer RiskScorer) ([]Alert, error) {
	if workers < 1 {
		return nil, errors.New("workers must be at least 1")
	}

	if threshold < 0 {
		return nil, errors.New("threshold must not be negative")
	}

	if scorer == nil {
		return nil, errors.New("scorer must not be nil")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Stage 1 owns event emission from the input slice.
	in := source(ctx, events)
	// Stage 2 fans scoring work out across a bounded worker set.
	out := parallelScore(ctx, workers, in, scorer)

	alerts := make([]Alert, 0, len(events))
	var firstErr error

	for result := range out {
		if result.err != nil && firstErr == nil {
			// The sink stage decides that one scorer failure aborts the pipeline.
			firstErr = fmt.Errorf("score checkout %s: %w", result.checkoutID, result.err)
			cancel()
			continue
		}

		if firstErr != nil || result.signal.Score < threshold {
			continue
		}

		alerts = append(alerts, buildAlert(result.signal))
	}

	if firstErr != nil {
		return nil, firstErr
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].CheckoutID < alerts[j].CheckoutID
	})

	return alerts, nil
}

type scoreResult struct {
	checkoutID string
	signal     RiskSignal
	err        error
}

func source(ctx context.Context, events []CheckoutEvent) <-chan CheckoutEvent {
	out := make(chan CheckoutEvent)

	go func() {
		defer close(out)

		for _, event := range events {
			select {
			case out <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

func parallelScore(ctx context.Context, workers int, in <-chan CheckoutEvent, scorer RiskScorer) <-chan scoreResult {
	// Buffering by worker count prevents the fan-in side from becoming an
	// accidental bottleneck when several scorers complete at once.
	out := make(chan scoreResult, workers)

	var workersWG sync.WaitGroup
	worker := func() {
		defer workersWG.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-in:
				if !ok {
					return
				}

				signal, err := scorer(ctx, event)

				select {
				case out <- scoreResult{checkoutID: event.ID, signal: signal, err: err}:
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
		workersWG.Wait()
		close(out)
	}()

	return out
}

func buildAlert(signal RiskSignal) Alert {
	severity := "high"
	if signal.Score >= 90 {
		severity = "critical"
	}

	return Alert{
		CheckoutID: signal.ID,
		Severity:   severity,
		Score:      signal.Score,
		Summary:    fmt.Sprintf("manual review required: %s", strings.Join(signal.Reasons, ", ")),
	}
}
