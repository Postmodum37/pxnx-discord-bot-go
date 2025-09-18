package queue

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"pxnx-discord-bot/music/types"
)

// createTestSource creates a test audio source with the given title
func createTestSource(title string) types.AudioSource {
	return types.AudioSource{
		Title:       title,
		URL:         "https://example.com/" + title,
		Duration:    "3:45",
		Thumbnail:   "https://example.com/thumb.jpg",
		Provider:    "test",
		RequestedBy: "user123",
		StreamURL:   "https://example.com/stream/" + title,
	}
}

func TestNewQueue(t *testing.T) {
	q := NewQueue()
	assert.NotNil(t, q)
	assert.True(t, q.IsEmpty())
	assert.Equal(t, 0, q.Size())
}

func TestQueueAdd(t *testing.T) {
	q := NewQueue()
	source1 := createTestSource("song1")
	source2 := createTestSource("song2")

	// Add first item
	q.Add(source1)
	assert.False(t, q.IsEmpty())
	assert.Equal(t, 1, q.Size())

	// Add second item
	q.Add(source2)
	assert.Equal(t, 2, q.Size())

	// Verify items are in correct order
	items := q.GetAll()
	assert.Equal(t, "song1", items[0].Title)
	assert.Equal(t, "song2", items[1].Title)
}

func TestQueueRemove(t *testing.T) {
	q := NewQueue()
	source1 := createTestSource("song1")
	source2 := createTestSource("song2")
	source3 := createTestSource("song3")

	q.Add(source1)
	q.Add(source2)
	q.Add(source3)

	// Remove middle item
	err := q.Remove(1)
	assert.NoError(t, err)
	assert.Equal(t, 2, q.Size())

	items := q.GetAll()
	assert.Equal(t, "song1", items[0].Title)
	assert.Equal(t, "song3", items[1].Title)

	// Test error cases
	err = q.Remove(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")

	err = q.Remove(10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

func TestQueueGet(t *testing.T) {
	q := NewQueue()
	source1 := createTestSource("song1")
	source2 := createTestSource("song2")

	q.Add(source1)
	q.Add(source2)

	// Get valid positions
	item, err := q.Get(0)
	assert.NoError(t, err)
	assert.Equal(t, "song1", item.Title)

	item, err = q.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, "song2", item.Title)

	// Queue should remain unchanged
	assert.Equal(t, 2, q.Size())

	// Test error cases
	_, err = q.Get(-1)
	assert.Error(t, err)

	_, err = q.Get(10)
	assert.Error(t, err)
}

func TestQueueGetAll(t *testing.T) {
	q := NewQueue()
	source1 := createTestSource("song1")
	source2 := createTestSource("song2")

	// Empty queue
	items := q.GetAll()
	assert.Empty(t, items)

	// Add items and test
	q.Add(source1)
	q.Add(source2)

	items = q.GetAll()
	assert.Len(t, items, 2)
	assert.Equal(t, "song1", items[0].Title)
	assert.Equal(t, "song2", items[1].Title)

	// Verify it's a copy (modifying result shouldn't affect queue)
	items[0].Title = "modified"
	originalItems := q.GetAll()
	assert.Equal(t, "song1", originalItems[0].Title)
}

func TestQueueClear(t *testing.T) {
	q := NewQueue()
	source1 := createTestSource("song1")
	source2 := createTestSource("song2")

	q.Add(source1)
	q.Add(source2)
	assert.Equal(t, 2, q.Size())

	q.Clear()
	assert.True(t, q.IsEmpty())
	assert.Equal(t, 0, q.Size())
	assert.Empty(t, q.GetAll())
}

func TestQueueShuffle(t *testing.T) {
	q := NewQueue()

	// Test with empty queue (should not panic)
	q.Shuffle()
	assert.True(t, q.IsEmpty())

	// Test with single item (should remain unchanged)
	source1 := createTestSource("song1")
	q.Add(source1)
	q.Shuffle()
	items := q.GetAll()
	assert.Len(t, items, 1)
	assert.Equal(t, "song1", items[0].Title)

	// Test with multiple items
	q.Clear()
	sources := []types.AudioSource{
		createTestSource("song1"),
		createTestSource("song2"),
		createTestSource("song3"),
		createTestSource("song4"),
		createTestSource("song5"),
	}

	for _, source := range sources {
		q.Add(source)
	}

	originalOrder := q.GetAll()
	q.Shuffle()
	shuffledOrder := q.GetAll()

	// Should have same number of items
	assert.Equal(t, len(originalOrder), len(shuffledOrder))

	// Should contain all the same items (just potentially in different order)
	originalTitles := make(map[string]bool)
	for _, item := range originalOrder {
		originalTitles[item.Title] = true
	}

	for _, item := range shuffledOrder {
		assert.True(t, originalTitles[item.Title], "Shuffled queue contains unexpected item: %s", item.Title)
	}

	// Note: We can't reliably test that the order actually changed due to randomness,
	// but we can test that shuffle doesn't break the queue structure
}

func TestQueueNext(t *testing.T) {
	q := NewQueue()

	// Test empty queue
	item, hasNext := q.Next()
	assert.Nil(t, item)
	assert.False(t, hasNext)

	// Add items and test FIFO behavior
	source1 := createTestSource("song1")
	source2 := createTestSource("song2")
	source3 := createTestSource("song3")

	q.Add(source1)
	q.Add(source2)
	q.Add(source3)

	// First next should return first item and reduce size
	item, hasNext = q.Next()
	assert.NotNil(t, item)
	assert.True(t, hasNext)
	assert.Equal(t, "song1", item.Title)
	assert.Equal(t, 2, q.Size())

	// Second next should return second item
	item, hasNext = q.Next()
	assert.NotNil(t, item)
	assert.True(t, hasNext)
	assert.Equal(t, "song2", item.Title)
	assert.Equal(t, 1, q.Size())

	// Third next should return third item
	item, hasNext = q.Next()
	assert.NotNil(t, item)
	assert.True(t, hasNext)
	assert.Equal(t, "song3", item.Title)
	assert.Equal(t, 0, q.Size())
	assert.True(t, q.IsEmpty())

	// Fourth next should return nil
	item, hasNext = q.Next()
	assert.Nil(t, item)
	assert.False(t, hasNext)
}

func TestQueueSizeAndIsEmpty(t *testing.T) {
	q := NewQueue()

	// Empty queue
	assert.Equal(t, 0, q.Size())
	assert.True(t, q.IsEmpty())

	// Add one item
	q.Add(createTestSource("song1"))
	assert.Equal(t, 1, q.Size())
	assert.False(t, q.IsEmpty())

	// Add more items
	q.Add(createTestSource("song2"))
	q.Add(createTestSource("song3"))
	assert.Equal(t, 3, q.Size())
	assert.False(t, q.IsEmpty())

	// Remove all items
	q.Clear()
	assert.Equal(t, 0, q.Size())
	assert.True(t, q.IsEmpty())
}

// TestQueueConcurrency tests thread safety of queue operations
func TestQueueConcurrency(t *testing.T) {
	q := NewQueue()
	const numGoroutines = 10
	const itemsPerGoroutine = 100

	var wg sync.WaitGroup

	// Add items concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				source := createTestSource(fmt.Sprintf("song-%d-%d", id, j))
				q.Add(source)
			}
		}(i)
	}

	wg.Wait()

	// Verify all items were added
	expectedSize := numGoroutines * itemsPerGoroutine
	assert.Equal(t, expectedSize, q.Size())

	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				q.GetAll()
				q.Size()
				q.IsEmpty()
			}
		}()
	}

	wg.Wait()

	// Queue should still be intact
	assert.Equal(t, expectedSize, q.Size())

	// Test concurrent Next() operations
	var nextWg sync.WaitGroup
	itemsRetrieved := make(chan types.AudioSource, expectedSize)

	for i := 0; i < numGoroutines; i++ {
		nextWg.Add(1)
		go func() {
			defer nextWg.Done()
			for {
				if item, hasNext := q.Next(); hasNext {
					itemsRetrieved <- *item
				} else {
					break
				}
			}
		}()
	}

	nextWg.Wait()
	close(itemsRetrieved)

	// Count retrieved items
	retrievedCount := 0
	for range itemsRetrieved {
		retrievedCount++
	}

	assert.Equal(t, expectedSize, retrievedCount)
	assert.True(t, q.IsEmpty())
}

// BenchmarkQueueAdd benchmarks the Add operation
func BenchmarkQueueAdd(b *testing.B) {
	q := NewQueue()
	source := createTestSource("benchmark-song")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Add(source)
	}
}

// BenchmarkQueueNext benchmarks the Next operation
func BenchmarkQueueNext(b *testing.B) {
	q := NewQueue()
	source := createTestSource("benchmark-song")

	// Pre-populate queue
	for i := 0; i < b.N; i++ {
		q.Add(source)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Next()
	}
}

// BenchmarkQueueShuffle benchmarks the Shuffle operation
func BenchmarkQueueShuffle(b *testing.B) {
	q := NewQueue()
	source := createTestSource("benchmark-song")

	// Pre-populate queue with 1000 items
	for i := 0; i < 1000; i++ {
		q.Add(source)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Shuffle()
	}
}
