// Copyright 2020 Authors of Hubble
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package container

import (
	"context"

	"github.com/cilium/hubble/pkg/api/v1"
)

// RingReader is a reader for a Ring container.
type RingReader struct {
	ring *Ring
	idx  uint64
	c    <-chan *v1.Event
	stop chan struct{}
}

// NewRingReader creates a new RingReader that starts reading the ring at the
// position given by start.
func NewRingReader(ring *Ring, start uint64) *RingReader {
	return &RingReader{
		ring: ring,
		idx:  start,
		stop: make(chan struct{}),
	}
}

// Previous reads the event at the current position and decrement the read
// position by one. If no event was read, ok is false.
func (r *RingReader) Previous() (event *v1.Event, ok bool) {
	e, ok := r.ring.read(r.idx)
	if ok {
		r.idx--
	}
	return e, ok
}

// Next reads the event at the current position and increment the read position
// by one. If no event was read, ok is false.
func (r *RingReader) Next() (event *v1.Event, ok bool) {
	e, ok := r.ring.read(r.idx)
	if ok {
		r.idx++
	}
	return e, ok
}

// NextFollow reads the event at the current position and increment the read
// position by one. If there are no more event to read, NextFollow blocks
// until the next event is added to the ring or the context is cancelled.
func (r *RingReader) NextFollow(ctx context.Context) *v1.Event {
	if r.c == nil {
		r.c = r.ring.readFrom(r.stop, r.idx)
	}
	select {
	case e := <-r.c:
		r.idx++
		return e
	case <-ctx.Done():
		return nil
	}
}
