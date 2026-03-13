package sched_test

import (
	"sync/atomic"
	"testing"

	"github.com/MostafaMagdSalama/vortex/internal/sched"
)

func TestScheduler(t *testing.T) {
	s := sched.New(4)

	var count atomic.Int64

	for i := 0; i < 100; i++ {
		s.Submit(func() {
			count.Add(1)
		})
	}

	s.Stop() // waits for all 100 tasks, then shuts down

	if count.Load() != 100 {
		t.Fatalf("expected 100, got %d", count.Load())
	}
}