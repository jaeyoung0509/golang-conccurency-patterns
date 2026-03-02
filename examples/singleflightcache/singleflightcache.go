package singleflightcache

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/singleflight"
)

type PriceQuote struct {
	SKU       string
	Currency  string
	UnitCents int
	Source    string
}

type Loader func(context.Context, string) (PriceQuote, error)

type PriceService struct {
	loader Loader

	group singleflight.Group

	mu    sync.RWMutex
	cache map[string]PriceQuote
}

func NewPriceService(loader Loader) (*PriceService, error) {
	if loader == nil {
		return nil, errors.New("loader must not be nil")
	}

	return &PriceService{
		loader: loader,
		cache:  make(map[string]PriceQuote),
	}, nil
}

func (service *PriceService) Get(ctx context.Context, sku string) (PriceQuote, error) {
	if sku == "" {
		return PriceQuote{}, errors.New("sku must not be empty")
	}

	if quote, ok := service.lookup(sku); ok {
		return quote, nil
	}

	result := service.group.DoChan(sku, func() (interface{}, error) {
		if quote, ok := service.lookup(sku); ok {
			return quote, nil
		}

		quote, err := service.loader(ctx, sku)
		if err != nil {
			return PriceQuote{}, fmt.Errorf("load quote for %s: %w", sku, err)
		}

		service.mu.Lock()
		service.cache[sku] = quote
		service.mu.Unlock()

		return quote, nil
	})

	select {
	case <-ctx.Done():
		return PriceQuote{}, ctx.Err()
	case item := <-result:
		if item.Err != nil {
			return PriceQuote{}, item.Err
		}

		quote, ok := item.Val.(PriceQuote)
		if !ok {
			return PriceQuote{}, errors.New("singleflight returned unexpected value type")
		}

		return quote, nil
	}
}

func (service *PriceService) Invalidate(sku string) {
	service.mu.Lock()
	delete(service.cache, sku)
	service.mu.Unlock()
	service.group.Forget(sku)
}

func (service *PriceService) lookup(sku string) (PriceQuote, bool) {
	service.mu.RLock()
	quote, ok := service.cache[sku]
	service.mu.RUnlock()
	return quote, ok
}
