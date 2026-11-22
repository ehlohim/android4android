// Package duplex provides an in-memory transport for synthetic PCM fixtures.
// It has no network, device, audio capture, or telephony dependencies.
package duplex

import (
	"context"
	"errors"
)

const (
	SampleRate      = 16000
	FrameMillis     = 20
	SamplesPerFrame = SampleRate * FrameMillis / 1000
	maxQueuedFrames = 64
)

// Frame is a 20 ms mono signed 16-bit PCM block, copied by value.
// Sequence is a fixture marker; the transport neither changes nor validates it.
type Frame struct {
	Sequence uint64
	Samples  [SamplesPerFrame]int16
}

// Endpoint sends to its peer and receives only from its peer. Use one
// writer and one reader per endpoint when frame ordering matters.
type Endpoint struct {
	lifetime context.Context
	tx       chan<- Frame
	rx       <-chan Frame
}

// Pair creates endpoints A and B, with a separate bounded FIFO per direction.
// Zero capacity gives synchronous handoff. Cancel lifetime to stop both
// directions; cancellation abandons frames not yet delivered. Pair starts
// no background goroutines.
func Pair(lifetime context.Context, capacity int) (a, b Endpoint, err error) {
	if lifetime == nil {
		return Endpoint{}, Endpoint{}, errors.New("duplex: nil lifetime context")
	}
	if capacity < 0 || capacity > maxQueuedFrames {
		return Endpoint{}, Endpoint{}, errors.New("duplex: queue capacity must be between 0 and 64")
	}
	atob := make(chan Frame, capacity)
	btoa := make(chan Frame, capacity)
	return Endpoint{lifetime, atob, btoa}, Endpoint{lifetime, btoa, atob}, nil
}

// Send queues f for the peer, or waits for space. Canceling ctx affects
// only this operation. With concurrent cancellation and available space,
// either success or cancellation is possible, as with a channel select.
func (e Endpoint) Send(ctx context.Context, f Frame) error {
	if err := e.ready(ctx); err != nil {
		return err
	}
	select {
	case <-e.lifetime.Done():
		return e.lifetime.Err()
	case <-ctx.Done():
		return ctx.Err()
	case e.tx <- f:
		return nil
	}
}

// Receive reads one frame, waiting if necessary. Canceling ctx affects
// only this operation. A dequeue racing cancellation may still succeed.
func (e Endpoint) Receive(ctx context.Context) (Frame, error) {
	if err := e.ready(ctx); err != nil {
		return Frame{}, err
	}
	select {
	case <-e.lifetime.Done():
		return Frame{}, e.lifetime.Err()
	case <-ctx.Done():
		return Frame{}, ctx.Err()
	case f := <-e.rx:
		return f, nil
	}
}

func (e Endpoint) ready(ctx context.Context) error {
	if e.lifetime == nil {
		return errors.New("duplex: uninitialized endpoint")
	}
	if ctx == nil {
		return errors.New("duplex: nil operation context")
	}
	if err := e.lifetime.Err(); err != nil {
		return err
	}
	return ctx.Err()
}

// Fixture generates a reproducible synthetic sample pattern using fixed
// arithmetic. Distinct seeds identify directions and can reveal cross-talk;
// this data is not speech and the transport does not play it.
func Fixture(seed uint32, sequence uint64) Frame {
	f := Frame{Sequence: sequence}
	state := uint64(seed)<<32 ^ sequence
	for i := range f.Samples {
		state = state*6364136223846793005 + 1442695040888963407
		f.Samples[i] = int16(state>>52) - 2048
	}
	return f
}
