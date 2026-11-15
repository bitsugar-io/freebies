package ttlcache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSharesInFlightFetchPerKey(t *testing.T) {
	c := New[int](time.Minute, 10)
	var calls atomic.Int32
	release := make(chan struct{})
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := c.Get(t.Context(), "k", func(context.Context) (int, error) {
				calls.Add(1)
				<-release
				return 7, nil
			})
			if err != nil || v != 7 {
				t.Errorf("got %d, %v", v, err)
			}
		}()
	}
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()
	if calls.Load() != 1 {
		t.Errorf("fetch ran %d times, want 1", calls.Load())
	}
}

// A slow fetch for one key must not block another key.
func TestDifferentKeysFetchConcurrently(t *testing.T) {
	c := New[string](time.Minute, 10)
	slowStarted, release := make(chan struct{}), make(chan struct{})
	go func() {
		_, _ = c.Get(t.Context(), "slow", func(context.Context) (string, error) {
			close(slowStarted)
			<-release
			return "s", nil
		})
	}()
	<-slowStarted
	done := make(chan struct{})
	go func() {
		defer close(done)
		v, _ := c.Get(t.Context(), "fast", func(context.Context) (string, error) { return "f", nil })
		if v != "f" {
			t.Errorf("got %q", v)
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("fast key blocked behind slow key")
	}
	close(release)
}

func TestTTLAndErrorsNotCached(t *testing.T) {
	c := New[int](time.Minute, 10)
	now := time.Now()
	c.now = func() time.Time { return now }
	n := 0
	fetch := func(context.Context) (int, error) { n++; return n, nil }
	a, _ := c.Get(t.Context(), "k", fetch)
	b, _ := c.Get(t.Context(), "k", fetch)
	if a != 1 || b != 1 {
		t.Errorf("within ttl: %d %d, want cached 1 1", a, b)
	}
	now = now.Add(2 * time.Minute)
	if v, _ := c.Get(t.Context(), "k", fetch); v != 2 {
		t.Errorf("after ttl: %d, want refetch 2", v)
	}
	boom := errors.New("boom")
	if _, err := c.Get(t.Context(), "e", func(context.Context) (int, error) { return 0, boom }); !errors.Is(err, boom) {
		t.Fatal(err)
	}
	if v, err := c.Get(t.Context(), "e", func(context.Context) (int, error) { return 9, nil }); err != nil || v != 9 {
		t.Errorf("error was cached: %d %v", v, err)
	}
}

func TestBoundedSize(t *testing.T) {
	c := New[int](time.Hour, 3)
	now := time.Now()
	c.now = func() time.Time { now = now.Add(time.Second); return now }
	for i := range 10 {
		_, _ = c.Get(t.Context(), fmt.Sprint(i), func(context.Context) (int, error) { return i, nil })
	}
	if c.Len() != 3 {
		t.Errorf("len = %d, want 3", c.Len())
	}
	// The newest survive.
	calls := 0
	_, _ = c.Get(t.Context(), "9", func(context.Context) (int, error) { calls++; return 9, nil })
	if calls != 0 {
		t.Error("newest entry was evicted")
	}
}

func TestWaiterHonorsContext(t *testing.T) {
	c := New[int](time.Minute, 10)
	started, release := make(chan struct{}), make(chan struct{})
	go func() {
		_, _ = c.Get(t.Context(), "k", func(context.Context) (int, error) { close(started); <-release; return 1, nil })
	}()
	<-started
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	if _, err := c.Get(ctx, "k", nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want deadline exceeded", err)
	}
	close(release)
}
