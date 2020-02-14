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
	"fmt"
	"testing"
	"time"

	"github.com/cilium/hubble/pkg/api/v1"

	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
)

type want struct {
	e  *v1.Event
	ok bool
}

func TestRingReader_Previous(t *testing.T) {
	ring := NewRing(15)
	for i := 0; i < 15; i++ {
		ring.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: int64(i)}})
	}
	tests := []struct {
		start uint64
		count int
		want  []want
	}{
		{
			start: 13,
			count: 1,
			want: []want{
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 13}}, true},
			},
		}, {
			start: 13,
			count: 2,
			want: []want{
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 13}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 12}}, true},
			},
		}, {
			start: 5,
			count: 5,
			want: []want{
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 5}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 4}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 3}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 2}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 1}}, true},
			},
		}, {
			start: 0,
			count: 1,
			want: []want{
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 0}}, true},
			},
		}, {
			start: 0,
			count: 2,
			want: []want{
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 0}}, true},
				{nil, true},
			},
		}, {
			start: 14,
			count: 1,
			want: []want{
				{nil, false},
			},
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("read %d, start at position %d", tt.count, tt.start)
		t.Run(name, func(t *testing.T) {
			reader := NewRingReader(ring, tt.start)
			var got []want
			for i := 0; i < tt.count; i++ {
				e, ok := reader.Previous()
				got = append(got, want{e, ok})
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRingReader_Next(t *testing.T) {
	ring := NewRing(15)
	for i := 0; i < 15; i++ {
		ring.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: int64(i)}})
	}

	tests := []struct {
		start uint64
		count int
		want  []want
	}{
		{
			start: 0,
			count: 1,
			want: []want{
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 0}}, true},
			},
		}, {
			start: 0,
			count: 2,
			want: []want{
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 0}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 1}}, true},
			},
		}, {
			start: 5,
			count: 5,
			want: []want{
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 5}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 6}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 7}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 8}}, true},
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 9}}, true},
			},
		}, {
			start: 13,
			count: 1,
			want: []want{
				{&v1.Event{Timestamp: &types.Timestamp{Seconds: 13}}, true},
			},
		}, {
			start: 14,
			count: 1,
			want: []want{
				{nil, false},
			},
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("read %d, start at position %d", tt.count, tt.start)
		t.Run(name, func(t *testing.T) {
			reader := NewRingReader(ring, tt.start)
			var got []want
			for i := 0; i < tt.count; i++ {
				e, ok := reader.Next()
				got = append(got, want{e, ok})
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRingReader_NextFollow(t *testing.T) {
	ring := NewRing(15)
	for i := 0; i < 15; i++ {
		ring.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: int64(i)}})
	}

	tests := []struct {
		start       uint64
		count       int
		want        []*v1.Event
		wantTimeout bool
	}{
		{
			start: 0,
			count: 1,
			want: []*v1.Event{
				{Timestamp: &types.Timestamp{Seconds: 0}},
			},
		}, {
			start: 0,
			count: 2,
			want: []*v1.Event{
				{Timestamp: &types.Timestamp{Seconds: 0}},
				{Timestamp: &types.Timestamp{Seconds: 1}},
			},
		}, {
			start: 5,
			count: 5,
			want: []*v1.Event{
				{Timestamp: &types.Timestamp{Seconds: 5}},
				{Timestamp: &types.Timestamp{Seconds: 6}},
				{Timestamp: &types.Timestamp{Seconds: 7}},
				{Timestamp: &types.Timestamp{Seconds: 8}},
				{Timestamp: &types.Timestamp{Seconds: 9}},
			},
		}, {
			start: 13,
			count: 1,
			want: []*v1.Event{
				{Timestamp: &types.Timestamp{Seconds: 13}},
			},
		}, {
			start:       14,
			count:       1,
			want:        []*v1.Event{nil},
			wantTimeout: true,
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("read %d, start at position %d, expect timeout=%t", tt.count, tt.start, tt.wantTimeout)
		t.Run(name, func(t *testing.T) {
			reader := NewRingReader(ring, tt.start)
			var timedOut bool
			var got []*v1.Event
			for i := 0; i < tt.count; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				got = append(got, reader.NextFollow(ctx))
				select {
				case <-ctx.Done():
					timedOut = true
				default:
				}
				cancel()
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantTimeout, timedOut)
		})
	}
}
