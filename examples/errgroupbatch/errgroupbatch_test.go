package errgroupbatch

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunBackfillsPreservesInputOrder(t *testing.T) {
	jobs := []TenantBackfill{
		{TenantID: "tenant-a", WindowDays: 30, Priority: "high"},
		{TenantID: "tenant-b", WindowDays: 7, Priority: "normal"},
		{TenantID: "tenant-c", WindowDays: 90, Priority: "high"},
	}

	summaries, err := RunBackfills(context.Background(), jobs, 3, func(_ context.Context, job TenantBackfill) (BackfillSummary, error) {
		time.Sleep(time.Duration(4-job.WindowDays/30) * 5 * time.Millisecond)

		return BackfillSummary{
			TenantID:         job.TenantID,
			ProcessedEvents:  job.WindowDays * 100,
			RepairedSegments: job.WindowDays / 3,
			DurationMinutes:  job.WindowDays / 2,
		}, nil
	})
	if err != nil {
		t.Fatalf("RunBackfills returned error: %v", err)
	}

	for index, summary := range summaries {
		if summary.TenantID != jobs[index].TenantID {
			t.Fatalf("summary %d mismatch: got %s want %s", index, summary.TenantID, jobs[index].TenantID)
		}
	}
}

func TestRunBackfillsRespectsLimit(t *testing.T) {
	jobs := make([]TenantBackfill, 18)
	for index := range jobs {
		jobs[index] = TenantBackfill{TenantID: "tenant", WindowDays: 14}
	}

	var inFlight atomic.Int32
	var maxInFlight atomic.Int32

	_, err := RunBackfills(context.Background(), jobs, 4, func(_ context.Context, job TenantBackfill) (BackfillSummary, error) {
		current := inFlight.Add(1)
		defer inFlight.Add(-1)

		for {
			seen := maxInFlight.Load()
			if current <= seen || maxInFlight.CompareAndSwap(seen, current) {
				break
			}
		}

		time.Sleep(10 * time.Millisecond)

		return BackfillSummary{TenantID: job.TenantID}, nil
	})
	if err != nil {
		t.Fatalf("RunBackfills returned error: %v", err)
	}

	if maxInFlight.Load() > 4 {
		t.Fatalf("limit exceeded: got %d", maxInFlight.Load())
	}
}

func TestRunBackfillsCancelsSiblingsOnError(t *testing.T) {
	cancelObserved := make(chan struct{}, 1)

	jobs := []TenantBackfill{
		{TenantID: "slow-a"},
		{TenantID: "broken"},
		{TenantID: "slow-b"},
	}

	_, err := RunBackfills(context.Background(), jobs, 3, func(ctx context.Context, job TenantBackfill) (BackfillSummary, error) {
		if job.TenantID == "broken" {
			return BackfillSummary{}, errors.New("warehouse export corrupted")
		}

		<-ctx.Done()
		select {
		case cancelObserved <- struct{}{}:
		default:
		}

		return BackfillSummary{}, ctx.Err()
	})
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	select {
	case <-cancelObserved:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected sibling job to observe cancellation")
	}
}
