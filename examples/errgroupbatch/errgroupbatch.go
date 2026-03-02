package errgroupbatch

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type TenantBackfill struct {
	TenantID   string
	WindowDays int
	Priority   string
}

type BackfillSummary struct {
	TenantID         string
	ProcessedEvents  int
	RepairedSegments int
	DurationMinutes  int
	RequiresFollowUp bool
}

type BackfillFunc func(context.Context, TenantBackfill) (BackfillSummary, error)

func RunBackfills(ctx context.Context, jobs []TenantBackfill, limit int, runFn BackfillFunc) ([]BackfillSummary, error) {
	if limit < 1 {
		return nil, errors.New("limit must be at least 1")
	}

	if runFn == nil {
		return nil, errors.New("runFn must not be nil")
	}

	if len(jobs) == 0 {
		return []BackfillSummary{}, nil
	}

	for _, job := range jobs {
		if job.TenantID == "" {
			return nil, errors.New("tenantID must not be empty")
		}
	}

	summaries := make([]BackfillSummary, len(jobs))

	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(limit)

	for index, job := range jobs {
		index := index
		job := job

		group.Go(func() error {
			summary, err := runFn(ctx, job)
			if err != nil {
				return fmt.Errorf("backfill tenant %s: %w", job.TenantID, err)
			}

			summaries[index] = summary
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return summaries, nil
}
