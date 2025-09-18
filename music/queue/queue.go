package queue

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"pxnx-discord-bot/music/types"
)

// SimpleQueue implements the Queue interface with thread-safe operations
type SimpleQueue struct {
	items []types.AudioSource
	mu    sync.RWMutex
}

// NewQueue creates a new empty queue
func NewQueue() *SimpleQueue {
	return &SimpleQueue{
		items: make([]types.AudioSource, 0),
	}
}

// Add adds an audio source to the end of the queue
func (q *SimpleQueue) Add(source types.AudioSource) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, source)
}

// Remove removes an item at the specified position (0-indexed)
func (q *SimpleQueue) Remove(position int) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if position < 0 || position >= len(q.items) {
		return fmt.Errorf("position %d out of range (queue size: %d)", position, len(q.items))
	}

	// Remove item at position
	q.items = append(q.items[:position], q.items[position+1:]...)
	return nil
}

// Get retrieves an item at the specified position without removing it
func (q *SimpleQueue) Get(position int) (*types.AudioSource, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if position < 0 || position >= len(q.items) {
		return nil, fmt.Errorf("position %d out of range (queue size: %d)", position, len(q.items))
	}

	return &q.items[position], nil
}

// GetAll returns a copy of all items in the queue
func (q *SimpleQueue) GetAll() []types.AudioSource {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]types.AudioSource, len(q.items))
	copy(result, q.items)
	return result
}

// Clear removes all items from the queue
func (q *SimpleQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = q.items[:0] // Clear slice but keep capacity
}

// Shuffle randomly reorders all items in the queue using crypto/rand for security
func (q *SimpleQueue) Shuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Fisher-Yates shuffle algorithm with crypto/rand
	for i := len(q.items) - 1; i > 0; i-- {
		// Generate cryptographically secure random number
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			// Fallback to simple swap if crypto/rand fails
			// This should rarely happen, but we need to handle it gracefully
			continue
		}
		j := int(n.Int64())
		q.items[i], q.items[j] = q.items[j], q.items[i]
	}
}

// Next removes and returns the first item in the queue (FIFO)
func (q *SimpleQueue) Next() (*types.AudioSource, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil, false
	}

	// Get the first item
	item := q.items[0]

	// Remove it from the queue
	q.items = q.items[1:]

	return &item, true
}

// Size returns the number of items in the queue
func (q *SimpleQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// IsEmpty checks if the queue has no items
func (q *SimpleQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items) == 0
}
