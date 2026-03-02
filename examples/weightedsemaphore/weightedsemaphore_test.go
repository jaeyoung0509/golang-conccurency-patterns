package weightedsemaphore

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRenderCatalogRespectsWeightedCapacity(t *testing.T) {
	jobs := []RenderJob{
		{AssetID: "asset-a", MemoryMB: 3, Pages: 12, Preset: "magazine"},
		{AssetID: "asset-b", MemoryMB: 2, Pages: 8, Preset: "catalog"},
		{AssetID: "asset-c", MemoryMB: 4, Pages: 16, Preset: "catalog"},
		{AssetID: "asset-d", MemoryMB: 1, Pages: 2, Preset: "thumb"},
	}

	var inFlight atomic.Int64
	var maxInFlight atomic.Int64

	artifacts, err := RenderCatalog(context.Background(), jobs, 5, func(_ context.Context, job RenderJob) (RenderArtifact, error) {
		current := inFlight.Add(job.MemoryMB)
		defer inFlight.Add(-job.MemoryMB)

		for {
			seen := maxInFlight.Load()
			if current <= seen || maxInFlight.CompareAndSwap(seen, current) {
				break
			}
		}

		time.Sleep(15 * time.Millisecond)

		return RenderArtifact{
			AssetID:        job.AssetID,
			OutputPath:     "/renders/" + job.AssetID + ".pdf",
			ThumbnailCount: job.Pages / 2,
			PeakMemoryMB:   job.MemoryMB,
		}, nil
	})
	if err != nil {
		t.Fatalf("RenderCatalog returned error: %v", err)
	}

	if len(artifacts) != len(jobs) {
		t.Fatalf("unexpected artifact count: got %d want %d", len(artifacts), len(jobs))
	}

	if maxInFlight.Load() > 5 {
		t.Fatalf("weighted capacity exceeded: got %d", maxInFlight.Load())
	}
}

func TestRenderCatalogCancelsOnError(t *testing.T) {
	cancelObserved := make(chan struct{}, 1)

	jobs := []RenderJob{
		{AssetID: "asset-a", MemoryMB: 2},
		{AssetID: "broken", MemoryMB: 2},
		{AssetID: "asset-b", MemoryMB: 1},
	}

	_, err := RenderCatalog(context.Background(), jobs, 5, func(ctx context.Context, job RenderJob) (RenderArtifact, error) {
		if job.AssetID == "broken" {
			return RenderArtifact{}, errors.New("renderer process crashed")
		}

		<-ctx.Done()
		select {
		case cancelObserved <- struct{}{}:
		default:
		}

		return RenderArtifact{}, ctx.Err()
	})
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	select {
	case <-cancelObserved:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected sibling render to observe cancellation")
	}
}

func TestRenderCatalogRejectsOversizedJobs(t *testing.T) {
	_, err := RenderCatalog(context.Background(), []RenderJob{
		{AssetID: "asset-a", MemoryMB: 8},
	}, 5, func(context.Context, RenderJob) (RenderArtifact, error) {
		return RenderArtifact{}, nil
	})
	if err == nil {
		t.Fatal("expected validation error but got nil")
	}
}
