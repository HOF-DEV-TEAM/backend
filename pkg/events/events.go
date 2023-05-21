package events

import (
	"context"
	"golang.org/x/sync/errgroup"
	"sync"
)

type EventHandler[V any] func(context.Context, V) error

type EventStore[V any, K comparable] map[K][]EventHandler[V]

type Event[V any, K comparable] struct {
	subs EventStore[V, K]
	mu   sync.RWMutex
}

func (g *Event[V, K]) Subscribe(k K, fn EventHandler[V]) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if fn == nil {
		panic("no callback is registered")
	}

	if g.subs == nil {
		g.subs = make(map[K][]EventHandler[V])
	}

	g.subs[k] = append(g.subs[k], fn)
}

type PublishOptions[T any, K comparable] struct {
	EventData T
	EventKey  K
	Context   context.Context
}

func (g *Event[V, K]) Publish(options *PublishOptions[V, K]) error {
	//handle event
	eg := new(errgroup.Group)
	if _, ok := g.subs[options.EventKey]; ok {
		for _, handle := range g.subs[options.EventKey] {
			eg.Go(func() error {
				err := handle(options.Context, options.EventData)

				if err != nil {
					return err
				}
				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

type Observable[V any, K comparable] struct {
	Options      *PublishOptions[V, K]
	EventEmitter Event[V, K]
}

func NewObservable[V any, K comparable](eventKey K) *Observable[V, K] {
	return &Observable[V, K]{
		Options: &PublishOptions[V, K]{
			EventKey: eventKey,
		},
	}
}

// Watch for changes to our value
func (ws *Observable[V, K]) Watch(fn EventHandler[V]) {
	ws.EventEmitter.Subscribe(ws.Options.EventKey, fn)
}

// Set a new value publishing to all subscribers
func (ws *Observable[T, K]) Set(ctx context.Context, value T) error {
	ws.Options.EventData = value
	ws.Options.Context = ctx
	return ws.EventEmitter.Publish(ws.Options)
}
