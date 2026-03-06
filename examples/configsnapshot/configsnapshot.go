package configsnapshot

import (
	"context"
	"errors"
	"maps"
	"sync"
	"sync/atomic"
)

var ErrNilLoader = errors.New("loader must not be nil")

type Snapshot struct {
	Version      int
	FeatureFlags map[string]bool
	RegionBudget map[string]int
}

type Loader func(context.Context) (Snapshot, error)

type Store struct {
	loader Loader

	bootstrapOnce sync.Once
	bootstrapErr  error

	reloadMu sync.Mutex
	current  atomic.Pointer[Snapshot]
}

func NewStore(loader Loader) (*Store, error) {
	if loader == nil {
		return nil, ErrNilLoader
	}

	return &Store{loader: loader}, nil
}

func (s *Store) Bootstrap(ctx context.Context) error {
	s.bootstrapOnce.Do(func() {
		s.bootstrapErr = s.Reload(ctx)
	})

	return s.bootstrapErr
}

func (s *Store) Reload(ctx context.Context) error {
	s.reloadMu.Lock()
	defer s.reloadMu.Unlock()

	next, err := s.loader(ctx)
	if err != nil {
		return err
	}

	cloned := cloneSnapshot(next)
	s.current.Store(&cloned)

	return nil
}

func (s *Store) Current() (Snapshot, bool) {
	ptr := s.current.Load()
	if ptr == nil {
		return Snapshot{}, false
	}

	return cloneSnapshot(*ptr), true
}

func (s *Store) Enabled(flag string) bool {
	ptr := s.current.Load()
	if ptr == nil {
		return false
	}

	return ptr.FeatureFlags[flag]
}

func (s *Store) RegionBudgetFor(region string) (int, bool) {
	ptr := s.current.Load()
	if ptr == nil {
		return 0, false
	}

	budget, ok := ptr.RegionBudget[region]
	return budget, ok
}

func cloneSnapshot(in Snapshot) Snapshot {
	return Snapshot{
		Version:      in.Version,
		FeatureFlags: maps.Clone(in.FeatureFlags),
		RegionBudget: maps.Clone(in.RegionBudget),
	}
}
