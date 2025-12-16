package events

import (
	"context"
	"sync"
)

// Handler is a function that handles events
type Handler func(ctx context.Context, event *Event) error

// Subscriber represents an event subscriber
type Subscriber struct {
	ID       string
	EventType EventType
	Handler  Handler
}

// Bus is an in-memory event bus
type Bus struct {
	subscribers map[EventType][]*Subscriber
	mu          sync.RWMutex
	eventChan   chan *Event
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewBus creates a new event bus
func NewBus(bufferSize int) *Bus {
	return &Bus{
		subscribers: make(map[EventType][]*Subscriber),
		eventChan:   make(chan *Event, bufferSize),
		stopChan:    make(chan struct{}),
	}
}

// Start starts the event bus
func (b *Bus) Start(ctx context.Context) {
	b.wg.Add(1)
	go b.processEvents(ctx)
}

// Stop stops the event bus gracefully
func (b *Bus) Stop() {
	close(b.stopChan)
	b.wg.Wait()
}

// Subscribe subscribes a handler to an event type
func (b *Bus) Subscribe(eventType EventType, handler Handler) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	subscriber := &Subscriber{
		ID:       generateSubscriberID(),
		EventType: eventType,
		Handler:  handler,
	}

	b.subscribers[eventType] = append(b.subscribers[eventType], subscriber)

	return subscriber.ID
}

// SubscribeAll subscribes a handler to all event types
func (b *Bus) SubscribeAll(handler Handler) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	subscriber := &Subscriber{
		ID:       generateSubscriberID(),
		EventType: "", // Empty string means all events
		Handler:  handler,
	}

	// Add to special "all events" subscription
	b.subscribers[""] = append(b.subscribers[""], subscriber)

	return subscriber.ID
}

// Unsubscribe removes a subscriber
func (b *Bus) Unsubscribe(subscriberID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Search all event types
	for eventType, subs := range b.subscribers {
		for i, sub := range subs {
			if sub.ID == subscriberID {
				// Remove subscriber from slice
				b.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
				return
			}
		}
	}
}

// Publish publishes an event to the bus
func (b *Bus) Publish(event *Event) {
	select {
	case b.eventChan <- event:
		// Event sent successfully
	case <-b.stopChan:
		// Bus is stopped, discard event
	default:
		// Channel full, discard event (non-blocking)
	}
}

// PublishSync publishes an event synchronously and waits for all handlers
func (b *Bus) PublishSync(ctx context.Context, event *Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Get subscribers for this event type
	subscribers := b.getSubscribers(event.Type)

	// Call all handlers synchronously
	for _, sub := range subscribers {
		if err := sub.Handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

// processEvents processes events from the channel
func (b *Bus) processEvents(ctx context.Context) {
	defer b.wg.Done()

	for {
		select {
		case event := <-b.eventChan:
			b.handleEvent(ctx, event)
		case <-b.stopChan:
			// Drain remaining events
			for {
				select {
				case event := <-b.eventChan:
					b.handleEvent(ctx, event)
				default:
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// handleEvent handles a single event
func (b *Bus) handleEvent(ctx context.Context, event *Event) {
	b.mu.RLock()
	subscribers := b.getSubscribers(event.Type)
	b.mu.RUnlock()

	// Call all handlers asynchronously
	var wg sync.WaitGroup
	for _, sub := range subscribers {
		wg.Add(1)
		go func(handler Handler) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					// Handler panicked, log but don't crash
				}
			}()
			_ = handler(ctx, event)
		}(sub.Handler)
	}

	wg.Wait()
}

// getSubscribers returns all subscribers for an event type
func (b *Bus) getSubscribers(eventType EventType) []*Subscriber {
	// Get specific subscribers
	subscribers := make([]*Subscriber, 0)
	if subs, exists := b.subscribers[eventType]; exists {
		subscribers = append(subscribers, subs...)
	}

	// Add "all events" subscribers
	if allSubs, exists := b.subscribers[""]; exists {
		subscribers = append(subscribers, allSubs...)
	}

	return subscribers
}

// GetSubscriberCount returns the number of subscribers for an event type
func (b *Bus) GetSubscriberCount(eventType EventType) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := 0
	if subs, exists := b.subscribers[eventType]; exists {
		count += len(subs)
	}

	// Add "all events" subscribers
	if allSubs, exists := b.subscribers[""]; exists {
		count += len(allSubs)
	}

	return count
}

// GetTotalSubscriberCount returns the total number of subscribers
func (b *Bus) GetTotalSubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := 0
	for _, subs := range b.subscribers {
		count += len(subs)
	}

	return count
}

// generateSubscriberID generates a unique subscriber ID
func generateSubscriberID() string {
	return generateEventID()
}
