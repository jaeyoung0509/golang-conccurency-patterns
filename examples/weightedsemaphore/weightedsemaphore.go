package weightedsemaphore

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type RenderJob struct {
	AssetID  string
	MemoryMB int64
	Pages    int
	Preset   string
}

type RenderArtifact struct {
	AssetID        string
	OutputPath     string
	ThumbnailCount int
	PeakMemoryMB   int64
}

type RenderFunc func(context.Context, RenderJob) (RenderArtifact, error)

func RenderCatalog(ctx context.Context, jobs []RenderJob, capacity int64, renderFn RenderFunc) ([]RenderArtifact, error) {
	if capacity < 1 {
		return nil, errors.New("capacity must be at least 1")
	}

	if renderFn == nil {
		return nil, errors.New("renderFn must not be nil")
	}

	if len(jobs) == 0 {
		return []RenderArtifact{}, nil
	}

	for _, job := range jobs {
		if job.AssetID == "" {
			return nil, errors.New("assetID must not be empty")
		}

		if job.MemoryMB < 1 {
			return nil, fmt.Errorf("job %s must request at least 1 MB", job.AssetID)
		}

		if job.MemoryMB > capacity {
			return nil, fmt.Errorf("job %s requires %d MB but capacity is %d MB", job.AssetID, job.MemoryMB, capacity)
		}
	}

	sem := semaphore.NewWeighted(capacity)
	artifacts := make([]RenderArtifact, len(jobs))

	group, ctx := errgroup.WithContext(ctx)

	for index, job := range jobs {
		index := index
		job := job

		group.Go(func() error {
			if err := sem.Acquire(ctx, job.MemoryMB); err != nil {
				return err
			}
			defer sem.Release(job.MemoryMB)

			artifact, err := renderFn(ctx, job)
			if err != nil {
				return fmt.Errorf("render asset %s: %w", job.AssetID, err)
			}

			artifacts[index] = artifact
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return artifacts, nil
}
