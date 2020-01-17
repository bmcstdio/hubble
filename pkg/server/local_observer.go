// Copyright 2019 Authors of Cilium
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

package server

import (
	"context"

	"github.com/cilium/cilium/pkg/math"
	"github.com/cilium/cilium/pkg/monitor"
	"github.com/cilium/cilium/pkg/monitor/api"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"go.uber.org/zap"

	pb "github.com/cilium/hubble/api/v1/flow"
	"github.com/cilium/hubble/api/v1/observer"
	"github.com/cilium/hubble/pkg/container"
	"github.com/cilium/hubble/pkg/logger"
	"github.com/cilium/hubble/pkg/parser"
)

// LocalObserverServer is an implementation of the server.Observer interface
// that's meant to be run embedded inside the Cilium process. It ignores all
// the state change events since the state is available locally.
type LocalObserverServer struct {
	// ring buffer that contains the references of all flows
	ring *container.Ring

	// events is the channel used by the writer(s) to send the flow data
	// into the observer server.
	events chan *pb.Payload

	// stopped is mostly used in unit tests to signalize when the events
	// channel is empty, once it's closed.
	stopped chan struct{}

	log *zap.Logger

	// channel to receive events from observer server.
	eventschan chan *observer.GetFlowsResponse

	// payloadParser decodes pb.Payload into pb.Flow
	payloadParser *parser.Parser

	// noop channels
	endpointEvents chan monitorAPI.AgentNotify
	logRecord      chan monitor.LogRecordNotify
}

// NewLocalServer returns a new local observer server.
func NewLocalServer(
	payloadParser *parser.Parser,
	maxFlows int,
) *LocalObserverServer {
	return &LocalObserverServer{
		log:  logger.GetLogger(),
		ring: container.NewRing(maxFlows),
		// have a channel with 1% of the max flows that we can receive
		events:         make(chan *pb.Payload, uint64(math.IntMin(maxFlows/100, 100))),
		stopped:        make(chan struct{}),
		eventschan:     make(chan *observer.GetFlowsResponse, 100),
		payloadParser:  payloadParser,
		endpointEvents: make(chan monitorAPI.AgentNotify),
		logRecord:      make(chan monitor.LogRecordNotify),
	}
}

// Start starts the server to handle the events sent to the events channel as
// well as handle events to the EpAdd and EpDel channels.
func (s *LocalObserverServer) Start() {
	go s.consumeEndpointEvents()
	go s.consumeLogRecordNotify()
	processEvents(s)
}

// GetEventsChannel returns the event channel to receive pb.Payload events.
func (s *LocalObserverServer) GetEventsChannel() chan *pb.Payload {
	return s.events
}

// GetRingBuffer implements Observer.GetRingBuffer.
func (s *LocalObserverServer) GetRingBuffer() *container.Ring {
	return s.ring
}

// GetLogger implements Observer.GetLogger.
func (s *LocalObserverServer) GetLogger() *zap.Logger {
	return s.log
}

// GetStopped implements Observer.GetStopped.
func (s *LocalObserverServer) GetStopped() chan struct{} {
	return s.stopped
}

// GetPayloadParser implements Observer.GetPayloadParser.
func (s *LocalObserverServer) GetPayloadParser() *parser.Parser {
	return s.payloadParser
}

// ServerStatus should have a comment, apparently. It returns the server status.
func (s *LocalObserverServer) ServerStatus(
	ctx context.Context, req *observer.ServerStatusRequest,
) (*observer.ServerStatusResponse, error) {
	return getServerStatusFromObserver(s)
}

// GetFlows implements the proto method for client requests.
func (s *LocalObserverServer) GetFlows(
	req *observer.GetFlowsRequest,
	server observer.Observer_GetFlowsServer,
) (err error) {
	return getFlowsFromObserver(req, server, s)
}

// GetEndpointEventsChannel implements Observer.GetEndpointEventsChannel.
func (s *LocalObserverServer) GetEndpointEventsChannel() chan<- api.AgentNotify {
	return s.endpointEvents
}

// GetLogRecordNotifyChannel implements Observer.GetLogRecordNotifyChannel.
func (s *LocalObserverServer) GetLogRecordNotifyChannel() chan<- monitor.LogRecordNotify {
	return s.logRecord
}

// StartMirroringIPCache implements Observer.StartMirroringIPCache.
func (s *LocalObserverServer) StartMirroringIPCache(ipCacheEvents <-chan api.AgentNotify) {
	go s.consumeAgentNotify(ipCacheEvents)
}

func (s *LocalObserverServer) consumeEndpointEvents() {
	for range s.endpointEvents {
	}
}

func (s *LocalObserverServer) consumeLogRecordNotify() {
	for range s.logRecord {
	}
}

func (s *LocalObserverServer) consumeAgentNotify(ipCacheEvents <-chan api.AgentNotify) {
	for range ipCacheEvents {
	}
}
