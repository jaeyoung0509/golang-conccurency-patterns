package requestreply

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrInvalidBuffer = errors.New("buffer must be at least 1")
	ErrNilEngine     = errors.New("engine must not be nil")
	ErrBrokerStopped = errors.New("broker stopped")
)

type FraudCheck struct {
	OrderID        string
	AmountCents    int
	CountryCode    string
	AccountAgeDays int
}

type Decision struct {
	OrderID  string
	Approved bool
	Queue    string
	Score    int
}

type EngineFunc func(context.Context, FraudCheck) (Decision, error)

type Broker struct {
	engine   EngineFunc
	requests chan request
	done     chan struct{}
	doneOnce sync.Once
	mu       sync.RWMutex
	closed   bool
}

type request struct {
	check FraudCheck
	// Each caller supplies its own reply path, which keeps response ownership explicit.
	reply chan result
}

type result struct {
	decision Decision
	err      error
}

func NewBroker(buffer int, engine EngineFunc) (*Broker, error) {
	if buffer < 1 {
		return nil, ErrInvalidBuffer
	}
	if engine == nil {
		return nil, ErrNilEngine
	}

	return &Broker{
		engine:   engine,
		requests: make(chan request, buffer),
		done:     make(chan struct{}),
	}, nil
}

func (b *Broker) Run(ctx context.Context) {
	defer func() {
		b.mu.Lock()
		b.closed = true
		b.mu.Unlock()

		b.doneOnce.Do(func() {
			close(b.done)
		})
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case req := <-b.requests:
			// The broker serializes engine access and routes the result back through
			// the caller-owned reply channel.
			decision, err := b.engine(ctx, req.check)

			select {
			case req.reply <- result{decision: decision, err: err}:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (b *Broker) Check(ctx context.Context, check FraudCheck) (Decision, error) {
	// Buffer 1 ensures the broker can finish replying even if the caller times out
	// after the request was already admitted.
	reply := make(chan result, 1)

	req := request{
		check: check,
		reply: reply,
	}

	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return Decision{}, ErrBrokerStopped
	}
	select {
	case b.requests <- req:
		b.mu.RUnlock()
	case <-ctx.Done():
		b.mu.RUnlock()
		return Decision{}, ctx.Err()
	case <-b.done:
		b.mu.RUnlock()
		return Decision{}, ErrBrokerStopped
	}

	select {
	case res := <-reply:
		return res.decision, res.err
	case <-ctx.Done():
		return Decision{}, ctx.Err()
	case <-b.done:
		return Decision{}, ErrBrokerStopped
	}
}
