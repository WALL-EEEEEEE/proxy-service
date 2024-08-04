package service

import (
	"context"
	"time"

	client "github.com/WALL-EEEEEEE/proxy-service/gateway/internal/client"
	"github.com/WALL-EEEEEEE/proxy-service/gateway/log"
)

type EventType string

const (
	EVENT_PROXY_BLOCKED     EventType = "blocked"
	EVENT_PROXY_PASSED      EventType = "passed"
	EVENT_PROXY_UNAVAILABLE EventType = "unavailable"
)

type Event interface {
	Event() EventType
}

type BlockedEvent struct {
	Time  time.Time     `json:"time"`
	Site  string        `json:"site"`
	Proxy string        `json:"proxy"`
	Cost  time.Duration `json:"cost"`
}

func (e BlockedEvent) Event() EventType {
	return EVENT_PROXY_BLOCKED
}

type PassedEvent struct {
	Time  time.Time     `json:"time"`
	Site  string        `json:"site"`
	Proxy string        `json:"proxy"`
	Cost  time.Duration `json:"cost"`
}

func (e PassedEvent) Event() EventType {
	return EVENT_PROXY_PASSED
}

type UnavailableEvent struct {
	Proxy string `json:"proxy"`
}

func (e UnavailableEvent) Event() EventType {
	return EVENT_PROXY_UNAVAILABLE
}

type GatewayServiceOptions struct {
	logger *log.Logger
	ctx    *context.Context
}

type GatewayServiceOption func(*GatewayServiceOptions)

func LogGatewayServiceOption(logger *log.Logger) GatewayServiceOption {
	return func(options *GatewayServiceOptions) {
		options.logger = logger
	}
}
func CtxGatewayServiceOption(ctx *context.Context) GatewayServiceOption {
	return func(options *GatewayServiceOptions) {
		options.ctx = ctx
	}
}

type GatewayService struct {
	events chan Event
	logger log.Logger
	client client.GatewayClient
}

func NewGatewayService(grpc_addr string, opts ...GatewayServiceOption) (*GatewayService, error) {
	options := &GatewayServiceOptions{}
	for _, opt := range opts {
		opt(options)
	}
	events := make(chan Event, 1)
	client, err := client.NewGatewayClient(grpc_addr, client.CtxGatewayClientOption(options.ctx), client.LogGatewayClientOption(options.logger))
	if err != nil {
		return nil, err
	}
	var logger log.Logger
	if options.logger != nil {
		logger = *options.logger
	} else {
		logger = log.DefaultLogger
	}
	service := &GatewayService{client: *client, events: events, logger: logger}
	service.tuneInEvents()
	return service, nil
}

func (s *GatewayService) CreateEvent(e Event) {
	s.events <- e
}

func (s *GatewayService) tuneInEvents() {
	go func() {
		for event := range s.events {
			s.logger.Infof("Recv: %s - %+v", event.Event(), event)
		}
	}()
}
