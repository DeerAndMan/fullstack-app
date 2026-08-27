package test

import (
	"sync"
	"testing"

	"fullstack-app/server/pkg/snowflake"
)

func TestSnowflakeNextIDUsesConfiguredNode(t *testing.T) {
	snowflake.Init(7)
	id := snowflake.NextID()
	if id <= 0 {
		t.Fatalf("NextID() = %d, want a positive ID", id)
	}
	if got, want := (id>>12)&1023, int64(7); got != want {
		t.Fatalf("node bits = %d, want %d", got, want)
	}

	snowflake.Init(2048)
	maskedID := snowflake.NextID()
	if got, want := (maskedID>>12)&1023, int64(0); got != want {
		t.Fatalf("masked node bits = %d, want %d", got, want)
	}
}

func TestSnowflakeNodeGeneratesUniqueIDsConcurrently(t *testing.T) {
	const (
		workers      = 8
		idsPerWorker = 100
		totalIDs     = workers * idsPerWorker
	)

	node := &snowflake.Node{}
	ids := make(chan int64, totalIDs)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerWorker; j++ {
				ids <- node.Generate()
			}
		}()
	}
	wg.Wait()
	close(ids)

	seen := make(map[int64]struct{}, totalIDs)
	for id := range ids {
		if id <= 0 {
			t.Fatalf("generated ID = %d, want a positive ID", id)
		}
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate generated ID: %d", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != totalIDs {
		t.Fatalf("unique ID count = %d, want %d", len(seen), totalIDs)
	}
}
