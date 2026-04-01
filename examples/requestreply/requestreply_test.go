package requestreply

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestBrokerRoutesRepliesToMatchingCaller(t *testing.T) {
	broker, err := NewBroker(8, func(_ context.Context, check FraudCheck) (Decision, error) {
		time.Sleep(time.Duration(len(check.OrderID)%3+1) * 5 * time.Millisecond)

		return Decision{
			OrderID:  check.OrderID,
			Approved: check.AmountCents < 20_000,
			Queue:    "standard",
			Score:    check.AccountAgeDays + check.AmountCents/100,
		}, nil
	})
	if err != nil {
		t.Fatalf("NewBroker returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broker.Run(ctx)

	checks := []FraudCheck{
		{OrderID: "ord-100", AmountCents: 12_000, AccountAgeDays: 120},
		{OrderID: "ord-200", AmountCents: 31_000, AccountAgeDays: 14},
		{OrderID: "ord-300", AmountCents: 6_500, AccountAgeDays: 300},
	}

	results := make(chan Decision, len(checks))
	errs := make(chan error, len(checks))

	var wg sync.WaitGroup
	for _, check := range checks {
		check := check
		wg.Add(1)

		go func() {
			defer wg.Done()

			decision, err := broker.Check(context.Background(), check)
			if err != nil {
				errs <- err
				return
			}

			results <- decision
		}()
	}

	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		t.Fatalf("broker.Check returned error: %v", err)
	}

	seen := make(map[string]Decision, len(checks))
	for decision := range results {
		seen[decision.OrderID] = decision
	}

	for _, check := range checks {
		decision, ok := seen[check.OrderID]
		if !ok {
			t.Fatalf("missing decision for %s", check.OrderID)
		}
		if decision.OrderID != check.OrderID {
			t.Fatalf("decision routed to wrong caller: got %s want %s", decision.OrderID, check.OrderID)
		}
	}
}

func TestBrokerDoesNotBlockIfCallerTimesOutAfterSend(t *testing.T) {
	release := make(chan struct{})

	broker, err := NewBroker(2, func(_ context.Context, check FraudCheck) (Decision, error) {
		if check.OrderID == "slow" {
			<-release
		}

		return Decision{OrderID: check.OrderID, Approved: true, Queue: "review"}, nil
	})
	if err != nil {
		t.Fatalf("NewBroker returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broker.Run(ctx)

	callerCtx, callerCancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer callerCancel()

	if _, err := broker.Check(callerCtx, FraudCheck{OrderID: "slow"}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("broker.Check error = %v, want deadline exceeded", err)
	}

	// A timed-out caller must not leave the broker wedged for the next request.
	close(release)

	decision, err := broker.Check(context.Background(), FraudCheck{OrderID: "fast"})
	if err != nil {
		t.Fatalf("broker.Check returned error after timeout path: %v", err)
	}
	if decision.OrderID != "fast" {
		t.Fatalf("unexpected follow-up decision: got %s", decision.OrderID)
	}
}

func TestCheckReturnsBrokerStoppedWhenRunContextEnds(t *testing.T) {
	broker, err := NewBroker(1, func(_ context.Context, check FraudCheck) (Decision, error) {
		return Decision{OrderID: check.OrderID}, nil
	})
	if err != nil {
		t.Fatalf("NewBroker returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		broker.Run(ctx)
	}()

	cancel()
	<-done

	if _, err := broker.Check(context.Background(), FraudCheck{OrderID: "ord-1"}); !errors.Is(err, ErrBrokerStopped) {
		t.Fatalf("broker.Check error = %v, want ErrBrokerStopped", err)
	}
}
