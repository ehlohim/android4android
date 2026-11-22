package duplex

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestFixtureDeterministic(t *testing.T) {
	f := Fixture(1, 7)
	if f != Fixture(1, 7) || f == Fixture(2, 7) || f.Samples == Fixture(1, 8).Samples {
		t.Fatal("fixtures must be reproducible and distinguish seed and sequence")
	}
	if f.Sequence != 7 || f.Samples == [SamplesPerFrame]int16{} {
		t.Fatal("fixture must carry its sequence and non-silent samples")
	}
}

func TestIndependentDirections(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	a, b, err := Pair(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	wantAB := Fixture(1, 0)
	f := wantAB
	if err := a.Send(ctx, f); err != nil {
		t.Fatal(err)
	}
	f.Samples[0]++ // local mutation cannot affect the enqueued frame

	// A-to-B is full, and B never drains it until reverse traffic succeeds.
	wantBA := Fixture(2, 0)
	if err := b.Send(ctx, wantBA); err != nil {
		t.Fatal(err)
	}
	requireFrame(t, ctx, a, wantBA)
	requireFrame(t, ctx, b, wantAB)

	// Independent queues preserve ordering in each direction over many frames.
	for sequence := uint64(1); sequence <= 50; sequence++ {
		ab, ba := Fixture(1, sequence), Fixture(2, sequence)
		if err := a.Send(ctx, ab); err != nil {
			t.Fatal(err)
		}
		if err := b.Send(ctx, ba); err != nil {
			t.Fatal(err)
		}
		requireFrame(t, ctx, a, ba)
		requireFrame(t, ctx, b, ab)
	}
}

func TestConcurrentHandoff(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	a, b, err := Pair(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}

	const count = 64
	done := make(chan error, 4)
	for i, e := range []Endpoint{a, b} {
		go func(e Endpoint, seed uint32) {
			for sequence := uint64(0); sequence < count; sequence++ {
				if err := e.Send(ctx, Fixture(seed, sequence)); err != nil {
					done <- err
					return
				}
			}
			done <- nil
		}(e, uint32(i+1))
		go func(e Endpoint, seed uint32) {
			for sequence := uint64(0); sequence < count; sequence++ {
				got, err := e.Receive(ctx)
				if err != nil {
					done <- err
					return
				}
				if got != Fixture(seed, sequence) {
					done <- fmt.Errorf("wrong direction or order: seed %d, sequence %d", seed, sequence)
					return
				}
			}
			done <- nil
		}(e, uint32(2-i))
	}
	for i := 0; i < cap(done); i++ {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal("concurrent send and receive did not finish")
		}
	}
}

func TestOperationCancellationKeepsSessionAlive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	a, b, err := Pair(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	queued := Fixture(1, 1)
	if err := a.Send(ctx, queued); err != nil {
		t.Fatal(err)
	}

	call, stopCall := context.WithCancel(ctx)
	started, done := make(chan struct{}), make(chan error, 1)
	go func() {
		done <- a.Send(observedContext{call, started}, Fixture(1, 2))
	}()
	waitReady(t, ctx, started)
	stopCall()
	requireCanceled(t, ctx, done)
	requireFrame(t, ctx, b, queued)

	// A canceled call must not take the next inbound frame.
	want := Fixture(2, 1)
	if err := b.Send(ctx, want); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Receive(call); !errors.Is(err, context.Canceled) {
		t.Fatalf("receive with canceled call context: %v", err)
	}
	requireFrame(t, ctx, a, want)
}

func TestLifetimeCancellationStopsOperations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lifetime, closeSession := context.WithCancel(ctx)
	defer closeSession()
	a, b, err := Pair(lifetime, 0)
	if err != nil {
		t.Fatal(err)
	}
	sendReady, receiveReady := make(chan struct{}), make(chan struct{})
	done := make(chan error, 2)
	go func() {
		done <- a.Send(observedContext{ctx, sendReady}, Fixture(1, 0))
	}()
	go func() {
		_, err := a.Receive(observedContext{ctx, receiveReady})
		done <- err
	}()
	waitReady(t, ctx, sendReady)
	waitReady(t, ctx, receiveReady)
	closeSession()
	requireCanceled(t, ctx, done)
	requireCanceled(t, ctx, done)
	if err := b.Send(ctx, Fixture(2, 0)); !errors.Is(err, context.Canceled) {
		t.Fatalf("peer send after session cancellation: %v", err)
	}
	if _, err := b.Receive(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("peer receive after session cancellation: %v", err)
	}
}

func requireFrame(t *testing.T, ctx context.Context, e Endpoint, want Frame) {
	t.Helper()
	got, err := e.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("received unexpected sample pattern (sequence %d, want %d)", got.Sequence, want.Sequence)
	}
}

func requireCanceled(t *testing.T, guard context.Context, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-guard.Done():
		t.Fatal("operation did not stop after cancellation")
	}
}

// A fresh observedContext is used once to signal when an operation evaluates
// its select clauses, without relying on sleeps or goroutine scheduling.
type observedContext struct {
	context.Context
	ready chan struct{}
}

func (c observedContext) Done() <-chan struct{} {
	close(c.ready)
	return c.Context.Done()
}

func waitReady(t *testing.T, guard context.Context, ready <-chan struct{}) {
	t.Helper()
	select {
	case <-ready:
	case <-guard.Done():
		t.Fatal("operation did not reach its wait")
	}
}
