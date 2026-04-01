package pipeline

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunAlertPipelineProducesSortedAlerts(t *testing.T) {
	events := []CheckoutEvent{
		{ID: "c-100", CustomerID: "cust-1", Region: "uk", TotalCents: 4200, NewDevice: false, PaymentFailures: 0},
		{ID: "c-300", CustomerID: "cust-2", Region: "cross-border", TotalCents: 128000, NewDevice: true, PaymentFailures: 3},
		{ID: "c-200", CustomerID: "cust-3", Region: "uk", TotalCents: 248000, NewDevice: true, PaymentFailures: 0},
	}

	alerts, err := RunAlertPipeline(context.Background(), events, 3, 70, func(_ context.Context, event CheckoutEvent) (RiskSignal, error) {
		score := 0
		reasons := make([]string, 0, 4)

		if event.TotalCents >= 200000 {
			score += 60
			reasons = append(reasons, "large basket value")
		}

		if event.PaymentFailures >= 3 {
			score += 45
			reasons = append(reasons, "repeated payment failures")
		} else if event.PaymentFailures >= 1 {
			score += 15
			reasons = append(reasons, "payment retry")
		}

		if event.NewDevice {
			score += 20
			reasons = append(reasons, "new device")
		}

		if event.Region == "cross-border" {
			score += 15
			reasons = append(reasons, "cross-border checkout")
		}

		return RiskSignal{
			CheckoutEvent: event,
			Score:         score,
			Reasons:       reasons,
		}, nil
	})
	if err != nil {
		t.Fatalf("RunAlertPipeline returned error: %v", err)
	}

	if len(alerts) != 2 {
		t.Fatalf("unexpected alert count: got %d want 2", len(alerts))
	}

	if alerts[0].CheckoutID != "c-200" || alerts[0].Severity != "high" {
		t.Fatalf("unexpected first alert: %#v", alerts[0])
	}

	if alerts[1].CheckoutID != "c-300" || alerts[1].Severity != "high" {
		t.Fatalf("unexpected second alert: %#v", alerts[1])
	}
}

func TestRunAlertPipelineCancelsWorkersOnError(t *testing.T) {
	cancelObserved := make(chan struct{}, 1)
	var started atomic.Int32

	events := []CheckoutEvent{
		{ID: "slow"},
		{ID: "broken"},
		{ID: "never"},
	}

	_, err := RunAlertPipeline(context.Background(), events, 2, 60, func(ctx context.Context, event CheckoutEvent) (RiskSignal, error) {
		started.Add(1)

		if event.ID == "broken" {
			return RiskSignal{}, errors.New("risk engine timeout")
		}

		// Healthy workers should notice the shared cancellation signal.
		<-ctx.Done()
		select {
		case cancelObserved <- struct{}{}:
		default:
		}

		return RiskSignal{}, ctx.Err()
	})
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	if started.Load() < 2 {
		t.Fatalf("expected at least two workers to start, got %d", started.Load())
	}

	select {
	case <-cancelObserved:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("pipeline worker did not observe cancellation")
	}
}
