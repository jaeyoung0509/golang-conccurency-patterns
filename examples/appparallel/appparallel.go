package appparallel

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

var (
	ErrLimitMustBePositive = errors.New("limit must be at least 1")
	ErrClosed              = errors.New("background group is closed")
)

type CheckoutRequest struct {
	UserID    string
	PartnerID string
	Amount    int
	Currency  string
}

type PartnerConfig struct {
	PartnerID       string
	SettlementDelay int
	FeeBPS          int
}

type RiskDecision struct {
	Approved bool
	Rule     string
}

type BalanceSnapshot struct {
	UserID     string
	Available  int
	HoldAmount int
}

type CheckoutSnapshot struct {
	UserID    string
	PartnerID string
	Partner   PartnerConfig
	Risk      RiskDecision
	Balance   BalanceSnapshot
}

type PartnerSummary struct {
	PartnerID string
	Tier      string
}

type PartnerConfigReader interface {
	LoadPartnerConfig(context.Context, string) (PartnerConfig, error)
}

type RiskChecker interface {
	EvaluateRisk(context.Context, CheckoutRequest) (RiskDecision, error)
}

type BalanceReader interface {
	LoadBalance(context.Context, string) (BalanceSnapshot, error)
}

type PartnerSummaryFetcher interface {
	FetchPartnerSummary(context.Context, string) (PartnerSummary, error)
}

type PartnerConfigReaderFunc func(context.Context, string) (PartnerConfig, error)

func (fn PartnerConfigReaderFunc) LoadPartnerConfig(ctx context.Context, partnerID string) (PartnerConfig, error) {
	return fn(ctx, partnerID)
}

type RiskCheckerFunc func(context.Context, CheckoutRequest) (RiskDecision, error)

func (fn RiskCheckerFunc) EvaluateRisk(ctx context.Context, req CheckoutRequest) (RiskDecision, error) {
	return fn(ctx, req)
}

type BalanceReaderFunc func(context.Context, string) (BalanceSnapshot, error)

func (fn BalanceReaderFunc) LoadBalance(ctx context.Context, userID string) (BalanceSnapshot, error) {
	return fn(ctx, userID)
}

type PartnerSummaryFetcherFunc func(context.Context, string) (PartnerSummary, error)

func (fn PartnerSummaryFetcherFunc) FetchPartnerSummary(ctx context.Context, partnerID string) (PartnerSummary, error) {
	return fn(ctx, partnerID)
}

type Dependencies struct {
	PartnerConfigs PartnerConfigReader
	Risk           RiskChecker
	Balances       BalanceReader
}

// LoadCheckoutSnapshot is the plain errgroup case:
// a single request, a finite set of sibling calls, and fail-fast cancellation.
func LoadCheckoutSnapshot(ctx context.Context, req CheckoutRequest, deps Dependencies) (CheckoutSnapshot, error) {
	if deps.PartnerConfigs == nil || deps.Risk == nil || deps.Balances == nil {
		return CheckoutSnapshot{}, errors.New("dependencies must not be nil")
	}

	group, groupCtx := errgroup.WithContext(ctx)

	var partner PartnerConfig
	var risk RiskDecision
	var balance BalanceSnapshot

	group.Go(func() error {
		result, err := deps.PartnerConfigs.LoadPartnerConfig(groupCtx, req.PartnerID)
		if err != nil {
			return fmt.Errorf("load partner config: %w", err)
		}
		partner = result
		return nil
	})

	group.Go(func() error {
		result, err := deps.Risk.EvaluateRisk(groupCtx, req)
		if err != nil {
			return fmt.Errorf("evaluate risk: %w", err)
		}
		risk = result
		return nil
	})

	group.Go(func() error {
		result, err := deps.Balances.LoadBalance(groupCtx, req.UserID)
		if err != nil {
			return fmt.Errorf("load balance: %w", err)
		}
		balance = result
		return nil
	})

	if err := group.Wait(); err != nil {
		return CheckoutSnapshot{}, err
	}

	return CheckoutSnapshot{
		UserID:    req.UserID,
		PartnerID: req.PartnerID,
		Partner:   partner,
		Risk:      risk,
		Balance:   balance,
	}, nil
}

// TaskGroup is the wrapped errgroup case: explicit Go(ctx) semantics and one
// place for limit policy.
type TaskGroup struct {
	ctx   context.Context
	group *errgroup.Group
}

func NewTaskGroup(ctx context.Context, limit int) (*TaskGroup, error) {
	if limit < 1 {
		return nil, ErrLimitMustBePositive
	}

	group, childCtx := errgroup.WithContext(ctx)
	group.SetLimit(limit)

	return &TaskGroup{
		ctx:   childCtx,
		group: group,
	}, nil
}

func (g *TaskGroup) Go(fn func(context.Context) error) {
	g.group.Go(func() error {
		return fn(g.ctx)
	})
}

func (g *TaskGroup) Wait() error {
	return g.group.Wait()
}

func LoadPartnerSummaries(ctx context.Context, partnerIDs []string, limit int, fetch PartnerSummaryFetcher) ([]PartnerSummary, error) {
	if fetch == nil {
		return nil, errors.New("fetch must not be nil")
	}

	group, err := NewTaskGroup(ctx, limit)
	if err != nil {
		return nil, err
	}

	results := make([]PartnerSummary, len(partnerIDs))
	for index, partnerID := range partnerIDs {
		index := index
		partnerID := partnerID

		group.Go(func(ctx context.Context) error {
			summary, err := fetch.FetchPartnerSummary(ctx, partnerID)
			if err != nil {
				return fmt.Errorf("fetch partner %s: %w", partnerID, err)
			}
			results[index] = summary
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// BackgroundGroup is the wrapped WaitGroup case: long-lived tasks with owned
// cancellation and shutdown semantics.
type BackgroundGroup struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
}

func NewBackgroundGroup(parent context.Context) *BackgroundGroup {
	if parent == nil {
		parent = context.Background()
	}

	ctx, cancel := context.WithCancel(parent)
	return &BackgroundGroup{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (g *BackgroundGroup) Go(fn func(context.Context)) error {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return ErrClosed
	}
	g.wg.Add(1)
	g.mu.Unlock()

	go func() {
		defer g.wg.Done()
		fn(g.ctx)
	}()

	return nil
}

func (g *BackgroundGroup) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	g.mu.Lock()
	if !g.closed {
		g.closed = true
		g.cancel()
	}
	g.mu.Unlock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		g.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
