package gracefulshutdown

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrClosed         = errors.New("processor is shut down")
	ErrInvalidQueue   = errors.New("queue size must be at least 1")
	ErrInvalidWorkers = errors.New("workers must be at least 1")
	ErrNilHandler     = errors.New("handler must not be nil")
)

type Event struct {
	OrderID string
	Kind    string
	Attempt int
}

type Handler func(Event) error

type Processor struct {
	handler Handler
	jobs    chan Event

	mu     sync.RWMutex
	closed bool
	wg     sync.WaitGroup
}

func NewProcessor(workers, queue int, handler Handler) (*Processor, error) {
	if workers < 1 {
		return nil, ErrInvalidWorkers
	}
	if queue < 1 {
		return nil, ErrInvalidQueue
	}
	if handler == nil {
		return nil, ErrNilHandler
	}

	p := &Processor{
		handler: handler,
		jobs:    make(chan Event, queue),
	}

	p.wg.Add(workers)
	for range workers {
		go p.worker()
	}

	return p, nil
}

func (p *Processor) worker() {
	defer p.wg.Done()

	for event := range p.jobs {
		_ = p.handler(event)
	}
}

func (p *Processor) Submit(ctx context.Context, event Event) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return ErrClosed
	}

	select {
	case p.jobs <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Processor) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.jobs)
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		p.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
