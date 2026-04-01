package contexttimeout

import (
	"context"
	"errors"
	"fmt"
)

type Profile struct {
	Name string
	Plan string
}

type UsageSnapshot struct {
	ActiveProjects int
	PendingJobs    int
}

type Recommendation struct {
	ID    string
	Title string
}

type Dashboard struct {
	UserID          string
	Profile         Profile
	Usage           UsageSnapshot
	Recommendations []Recommendation
}

type ProfileLoader func(context.Context, string) (Profile, error)
type UsageLoader func(context.Context, string) (UsageSnapshot, error)
type RecommendationLoader func(context.Context, string) ([]Recommendation, error)

type DashboardService struct {
	LoadProfile         ProfileLoader
	LoadUsage           UsageLoader
	LoadRecommendations RecommendationLoader
}

func (service DashboardService) Build(ctx context.Context, userID string) (Dashboard, error) {
	if userID == "" {
		return Dashboard{}, errors.New("userID must not be empty")
	}

	if service.LoadProfile == nil || service.LoadUsage == nil || service.LoadRecommendations == nil {
		return Dashboard{}, errors.New("all loaders must be configured")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		profile         *Profile
		usage           *UsageSnapshot
		recommendations []Recommendation
		err             error
	}

	// One slot per sibling call lets each goroutine report exactly once without
	// depending on the collector being scheduled immediately.
	results := make(chan result, 3)

	go func() {
		profile, err := service.LoadProfile(ctx, userID)
		if err != nil {
			results <- result{err: fmt.Errorf("load profile: %w", err)}
			return
		}

		results <- result{profile: &profile}
	}()

	go func() {
		usage, err := service.LoadUsage(ctx, userID)
		if err != nil {
			results <- result{err: fmt.Errorf("load usage: %w", err)}
			return
		}

		results <- result{usage: &usage}
	}()

	go func() {
		recommendations, err := service.LoadRecommendations(ctx, userID)
		if err != nil {
			results <- result{err: fmt.Errorf("load recommendations: %w", err)}
			return
		}

		results <- result{recommendations: recommendations}
	}()

	dashboard := Dashboard{UserID: userID}

	for range 3 {
		select {
		case <-ctx.Done():
			return Dashboard{}, ctx.Err()
		case result := <-results:
			if result.err != nil {
				// First failure cancels the rest of the request-scoped work.
				cancel()
				return Dashboard{}, result.err
			}

			if result.profile != nil {
				dashboard.Profile = *result.profile
			}

			if result.usage != nil {
				dashboard.Usage = *result.usage
			}

			if result.recommendations != nil {
				dashboard.Recommendations = result.recommendations
			}
		}
	}

	return dashboard, nil
}
