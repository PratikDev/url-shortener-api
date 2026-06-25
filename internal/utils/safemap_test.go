package utils

import (
	"log"
	"sync"
	"testing"
)

func TestSafeMapUpdate(t *testing.T) {
	safemap := NewSafeMap[int]()
	loop := 1000
	var wg sync.WaitGroup

	for range loop {
		wg.Go(func() {
			safemap.Update("key", func(current int, _ bool) int {
				return current + 1
			})
		})
	}

	wg.Wait()
	log.Printf("[SafeMap.Update] test finished")

	val, exists := safemap.Get("key")
	failed := !exists || val != loop
	if failed {
		t.Errorf("expected: %d, got: %d (exists: %v)", loop, val, exists)
	}
}