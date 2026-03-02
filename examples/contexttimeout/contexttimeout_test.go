package contexttimeout

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDashboardServiceBuildsCompositeResponse(t *testing.T) {
	service := DashboardService{
		LoadProfile: func(context.Context, string) (Profile, error) {
			time.Sleep(5 * time.Millisecond)
			return Profile{Name: "Alicia", Plan: "team"}, nil
		},
		LoadUsage: func(context.Context, string) (UsageSnapshot, error) {
			time.Sleep(10 * time.Millisecond)
			return UsageSnapshot{ActiveProjects: 7, PendingJobs: 2}, nil
		},
		LoadRecommendations: func(context.Context, string) ([]Recommendation, error) {
			return []Recommendation{
				{ID: "rec-1", Title: "Scale worker pool for nightly imports"},
			}, nil
		},
	}

	dashboard, err := service.Build(context.Background(), "user-42")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if dashboard.Profile.Name != "Alicia" || dashboard.Usage.ActiveProjects != 7 {
		t.Fatalf("unexpected dashboard: %#v", dashboard)
	}
}

func TestDashboardServiceCancelsSiblingCallsAfterFailure(t *testing.T) {
	cancelObserved := make(chan struct{}, 2)

	service := DashboardService{
		LoadProfile: func(context.Context, string) (Profile, error) {
			return Profile{}, errors.New("profile service unavailable")
		},
		LoadUsage: func(ctx context.Context, _ string) (UsageSnapshot, error) {
			<-ctx.Done()
			cancelObserved <- struct{}{}
			return UsageSnapshot{}, ctx.Err()
		},
		LoadRecommendations: func(ctx context.Context, _ string) ([]Recommendation, error) {
			<-ctx.Done()
			cancelObserved <- struct{}{}
			return nil, ctx.Err()
		},
	}

	_, err := service.Build(context.Background(), "user-42")
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()

	count := 0
	for count < 2 {
		select {
		case <-cancelObserved:
			count++
		case <-timer.C:
			t.Fatal("sibling calls did not observe cancellation")
		}
	}
}

func TestDashboardServiceRespectsDeadline(t *testing.T) {
	service := DashboardService{
		LoadProfile: func(ctx context.Context, _ string) (Profile, error) {
			<-ctx.Done()
			return Profile{}, ctx.Err()
		},
		LoadUsage: func(context.Context, string) (UsageSnapshot, error) {
			return UsageSnapshot{ActiveProjects: 7}, nil
		},
		LoadRecommendations: func(context.Context, string) ([]Recommendation, error) {
			return []Recommendation{}, nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := service.Build(ctx, "user-42")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
