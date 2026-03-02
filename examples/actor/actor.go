package actor

import (
	"context"
	"errors"
	"sync"
)

var ErrActorStopped = errors.New("actor stopped")

type StockSnapshot struct {
	SKU       string
	OnHand    int
	Reserved  int
	Available int
}

type InventoryActor struct {
	commands chan commandEnvelope
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

type inventoryState struct {
	sku          string
	onHand       int
	reservations map[string]int
}

type commandResult struct {
	snapshot StockSnapshot
	err      error
}

type actorCommand interface {
	run(*inventoryState) commandResult
}

type commandEnvelope struct {
	command actorCommand
	reply   chan commandResult
}

type reserveCommand struct {
	orderID  string
	quantity int
}

type releaseCommand struct {
	orderID  string
	quantity int
}

type restockCommand struct {
	quantity int
}

type snapshotCommand struct{}

func StartInventoryActor(sku string, initialStock int) (*InventoryActor, error) {
	if sku == "" {
		return nil, errors.New("sku must not be empty")
	}

	if initialStock < 0 {
		return nil, errors.New("initialStock must not be negative")
	}

	actor := &InventoryActor{
		commands: make(chan commandEnvelope),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	state := inventoryState{
		sku:          sku,
		onHand:       initialStock,
		reservations: make(map[string]int),
	}

	go actor.loop(&state)

	return actor, nil
}

func (actor *InventoryActor) Reserve(ctx context.Context, orderID string, quantity int) (StockSnapshot, error) {
	return actor.request(ctx, reserveCommand{orderID: orderID, quantity: quantity})
}

func (actor *InventoryActor) Release(ctx context.Context, orderID string, quantity int) (StockSnapshot, error) {
	return actor.request(ctx, releaseCommand{orderID: orderID, quantity: quantity})
}

func (actor *InventoryActor) Restock(ctx context.Context, quantity int) (StockSnapshot, error) {
	return actor.request(ctx, restockCommand{quantity: quantity})
}

func (actor *InventoryActor) Snapshot(ctx context.Context) (StockSnapshot, error) {
	return actor.request(ctx, snapshotCommand{})
}

func (actor *InventoryActor) Stop() {
	actor.stopOnce.Do(func() {
		close(actor.stop)
	})

	<-actor.done
}

func (actor *InventoryActor) loop(state *inventoryState) {
	defer close(actor.done)

	for {
		select {
		case <-actor.stop:
			return
		case envelope := <-actor.commands:
			envelope.reply <- envelope.command.run(state)
		}
	}
}

func (actor *InventoryActor) request(ctx context.Context, command actorCommand) (StockSnapshot, error) {
	reply := make(chan commandResult, 1)
	envelope := commandEnvelope{
		command: command,
		reply:   reply,
	}

	select {
	case <-ctx.Done():
		return StockSnapshot{}, ctx.Err()
	case <-actor.done:
		return StockSnapshot{}, ErrActorStopped
	case actor.commands <- envelope:
	}

	select {
	case <-ctx.Done():
		return StockSnapshot{}, ctx.Err()
	case <-actor.done:
		return StockSnapshot{}, ErrActorStopped
	case result := <-reply:
		return result.snapshot, result.err
	}
}

func (command reserveCommand) run(state *inventoryState) commandResult {
	if command.orderID == "" {
		return commandResult{err: errors.New("orderID must not be empty")}
	}

	if command.quantity <= 0 {
		return commandResult{err: errors.New("quantity must be positive")}
	}

	if state.available() < command.quantity {
		return commandResult{
			snapshot: state.snapshot(),
			err:      errors.New("insufficient stock"),
		}
	}

	state.reservations[command.orderID] += command.quantity

	return commandResult{snapshot: state.snapshot()}
}

func (command releaseCommand) run(state *inventoryState) commandResult {
	if command.orderID == "" {
		return commandResult{err: errors.New("orderID must not be empty")}
	}

	if command.quantity <= 0 {
		return commandResult{err: errors.New("quantity must be positive")}
	}

	current := state.reservations[command.orderID]
	if current < command.quantity {
		return commandResult{
			snapshot: state.snapshot(),
			err:      errors.New("release quantity exceeds reservation"),
		}
	}

	if current == command.quantity {
		delete(state.reservations, command.orderID)
	} else {
		state.reservations[command.orderID] = current - command.quantity
	}

	return commandResult{snapshot: state.snapshot()}
}

func (command restockCommand) run(state *inventoryState) commandResult {
	if command.quantity <= 0 {
		return commandResult{err: errors.New("quantity must be positive")}
	}

	state.onHand += command.quantity

	return commandResult{snapshot: state.snapshot()}
}

func (snapshotCommand) run(state *inventoryState) commandResult {
	return commandResult{snapshot: state.snapshot()}
}

func (state *inventoryState) available() int {
	return state.onHand - state.reserved()
}

func (state *inventoryState) reserved() int {
	total := 0
	for _, quantity := range state.reservations {
		total += quantity
	}

	return total
}

func (state *inventoryState) snapshot() StockSnapshot {
	reserved := state.reserved()

	return StockSnapshot{
		SKU:       state.sku,
		OnHand:    state.onHand,
		Reserved:  reserved,
		Available: state.onHand - reserved,
	}
}
