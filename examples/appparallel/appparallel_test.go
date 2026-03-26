package appparallel

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoadCheckoutSnapshotSuccess(t *testing.T) {
	deps := Dependencies{
		PartnerConfigs: PartnerConfigReaderFunc(func(_ context.Context, partnerID string) (PartnerConfig, error) {
			return PartnerConfig{
				PartnerID:       partnerID,
				SettlementDelay: 2,
				FeeBPS:          35,
			}, nil
		}),
		Risk: RiskCheckerFunc(func(_ context.Context, req CheckoutRequest) (RiskDecision, error) {
			return RiskDecision{
				Approved: req.Amount < 100000,
				Rule:     "checkout.default",
			}, nil
		}),
		Balances: BalanceReaderFunc(func(_ context.Context, userID string) (BalanceSnapshot, error) {
			return BalanceSnapshot{
				UserID:     userID,
				Available:  250000,
				HoldAmount: 10000,
			}, nil
		}),
	}

	snapshot, err := LoadCheckoutSnapshot(context.Background(), CheckoutRequest{
		UserID:    "user-7",
		PartnerID: "partner-1",
		Amount:    45000,
		Currency:  "GBP",
	}, deps)
	if err != nil {
		t.Fatalf("LoadCheckoutSnapshot returned error: %v", err)
	}

	if snapshot.Partner.PartnerID != "partner-1" {
		t.Fatalf("unexpected partner config: %+v", snapshot.Partner)
	}
	if !snapshot.Risk.Approved {
		t.Fatalf("expected approved risk decision, got %+v", snapshot.Risk)
	}
	if snapshot.Balance.UserID != "user-7" {
		t.Fatalf("unexpected balance snapshot: %+v", snapshot.Balance)
	}
}

func TestLoadCheckoutSnapshotCancelsSiblingsOnError(t *testing.T) {
	cancelObserved := make(chan struct{}, 1)

	deps := Dependencies{
		PartnerConfigs: PartnerConfigReaderFunc(func(ctx context.Context, _ string) (PartnerConfig, error) {
			<-ctx.Done()
			select {
			case cancelObserved <- struct{}{}:
			default:
			}
			return PartnerConfig{}, ctx.Err()
		}),
		Risk: RiskCheckerFunc(func(_ context.Context, _ CheckoutRequest) (RiskDecision, error) {
			return RiskDecision{}, errors.New("fraud backend timed out")
		}),
		Balances: BalanceReaderFunc(func(_ context.Context, userID string) (BalanceSnapshot, error) {
			return BalanceSnapshot{UserID: userID, Available: 120000}, nil
		}),
	}

	_, err := LoadCheckoutSnapshot(context.Background(), CheckoutRequest{
		UserID:    "user-9",
		PartnerID: "partner-9",
		Amount:    70000,
	}, deps)
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	select {
	case <-cancelObserved:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected sibling call to observe cancellation")
	}
}

func TestLoadPartnerSummariesRespectsLimitAndPreservesOrder(t *testing.T) {
	partnerIDs := []string{
		"partner-a",
		"partner-b",
		"partner-c",
		"partner-d",
		"partner-e",
		"partner-f",
	}

	var inFlight atomic.Int32
	var maxInFlight atomic.Int32

	summaries, err := LoadPartnerSummaries(context.Background(), partnerIDs, 2, PartnerSummaryFetcherFunc(func(_ context.Context, partnerID string) (PartnerSummary, error) {
		current := inFlight.Add(1)
		defer inFlight.Add(-1)

		for {
			seen := maxInFlight.Load()
			if current <= seen || maxInFlight.CompareAndSwap(seen, current) {
				break
			}
		}

		time.Sleep(10 * time.Millisecond)

		return PartnerSummary{
			PartnerID: partnerID,
			Tier:      "standard",
		}, nil
	}))
	if err != nil {
		t.Fatalf("LoadPartnerSummaries returned error: %v", err)
	}

	if maxInFlight.Load() > 2 {
		t.Fatalf("limit exceeded: got %d", maxInFlight.Load())
	}

	for index, summary := range summaries {
		if summary.PartnerID != partnerIDs[index] {
			t.Fatalf("summary %d mismatch: got %s want %s", index, summary.PartnerID, partnerIDs[index])
		}
	}
}

func TestBackgroundGroupShutdownCancelsTasksAndRejectsNewWork(t *testing.T) {
	group := NewBackgroundGroup(context.Background())

	stopped := make(chan struct{})
	if err := group.Go(func(ctx context.Context) {
		<-ctx.Done()
		close(stopped)
	}); err != nil {
		t.Fatalf("Go returned error: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := group.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	select {
	case <-stopped:
	default:
		t.Fatal("expected background task to stop")
	}

	if err := group.Go(func(context.Context) {}); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed, got %v", err)
	}
}

func TestBackgroundGroupShutdownHonorsCallerDeadline(t *testing.T) {
	group := NewBackgroundGroup(context.Background())
	done := make(chan struct{})

	if err := group.Go(func(context.Context) {
		defer close(done)
		time.Sleep(100 * time.Millisecond)
	}); err != nil {
		t.Fatalf("Go returned error: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := group.Shutdown(shutdownCtx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected background task to finish eventually")
	}
}
