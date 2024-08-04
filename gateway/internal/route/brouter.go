package route

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/gateway/internal/selector"
	service "github.com/WALL-EEEEEEE/proxy-service/gateway/internal/service"
	"github.com/WALL-EEEEEEE/proxy-service/gateway/log"

	manager_model "github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/sirupsen/logrus"
)

const (
	defaultRouteTableSize         = 20
	defaultRouteTableCap          = 10000
	defaultFallbackRouteTableSize = 10
	defaultFallbackRouteTableCap  = 20
)

var (
	ErrNoCallbackSet = fmt.Errorf("no callback set")
	ErrNoFallbackSet = fmt.Errorf("no fallback set")
	ErrNoRoute       = fmt.Errorf("no route")
)

type RouteSelector selector.Selector[Route[manager_model.Proxy]]

type ProxyBrouterOptions struct {
	logger      *log.Logger
	fb_tbl_size *int
	fb_tbl_cap  *int
	tbl_size    *int
	tbl_cap     *int
	selector    RouteSelector
}

type ProxyBrouterOption func(*ProxyBrouterOptions)

func LogProxyBrouterOption(logger *log.Logger) ProxyBrouterOption {
	return func(options *ProxyBrouterOptions) {
		options.logger = logger
	}
}
func RouteTableSizeProxyBrouterOption(tbl_size int) ProxyBrouterOption {
	return func(options *ProxyBrouterOptions) {
		options.tbl_size = &tbl_size
	}
}
func FallbackRouteTableCapProxyBrouterOption(tbl_cap int) ProxyBrouterOption {
	return func(options *ProxyBrouterOptions) {
		options.fb_tbl_cap = &tbl_cap
	}
}
func FallbackRouteTableSizeProxyBrouterOption(tbl_size int) ProxyBrouterOption {
	return func(options *ProxyBrouterOptions) {
		options.fb_tbl_size = &tbl_size
	}
}
func RouteTableCapProxyBrouterOption(tbl_cap int) ProxyBrouterOption {
	return func(options *ProxyBrouterOptions) {
		options.tbl_cap = &tbl_cap
	}
}
func SelectorProxyBrouterOption(selector RouteSelector) ProxyBrouterOption {
	return func(options *ProxyBrouterOptions) {
		options.selector = selector
	}
}
func MaxRetryRouteOption(retry int) RouteOption {
	return func(options *RouteOptions) {
		options.max_retry = &retry
	}
}

type ProxyBrouter struct {
	logger           log.Logger
	addr             string
	ctx              context.Context
	proxy_serv       service.ProxyService
	gateway_serv     service.GatewayService
	dyn_route_tbl    *RouteTable[manager_model.Proxy] //route table for proxies
	dyn_fb_route_tbl *RouteTable[manager_model.Proxy] //route table for fallback proxies
	tbl_size         int
	tbl_cap          int
	fb_tbl_size      int
	fb_tbl_cap       int
	selector         RouteSelector
}

func NewProxyBrouter(ctx context.Context, addr string, opts ...ProxyBrouterOption) (*ProxyBrouter, error) {
	options := &ProxyBrouterOptions{}
	for _, opt := range opts {
		opt(options)
	}
	s := ProxyBrouter{ctx: ctx, addr: addr}
	if options.logger != nil {
		s.logger = *options.logger
	} else {
		s.logger = log.DefaultLogger
	}
	if options.tbl_size != nil {
		s.tbl_size = *options.tbl_size
	} else {
		s.tbl_size = defaultRouteTableSize
	}
	if options.tbl_cap != nil {
		s.tbl_cap = *options.tbl_cap
	} else {
		s.tbl_cap = defaultRouteTableCap
	}
	if options.fb_tbl_size != nil {
		s.fb_tbl_size = *options.fb_tbl_size
	} else {
		s.fb_tbl_size = defaultFallbackRouteTableSize
	}
	if options.fb_tbl_cap != nil {
		s.fb_tbl_cap = *options.fb_tbl_cap
	} else {
		s.fb_tbl_cap = defaultFallbackRouteTableCap
	}
	if options.selector == nil {
		s.selector = selector.NewRoundRobin[Route[manager_model.Proxy]]()
	} else {
		s.selector = options.selector
	}
	err := s.init()
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *ProxyBrouter) init() error {
	var err error
	err = s.initService()
	if err != nil {
		return err
	}
	err = s.initRouteTbl()
	if err != nil {
		return err
	}
	return nil
}

func (s *ProxyBrouter) initService() error {
	proxy_serv, err := service.NewProxyService(s.addr, service.CtxProxyServiceOption(&s.ctx), service.PrefetchIntervalProxyServiceOption(time.Duration(60)*time.Second), service.LogProxyServiceOption(&s.logger))
	if err != nil {
		return err
	}
	s.proxy_serv = *proxy_serv
	gateway_serv, err := service.NewGatewayService(s.addr, service.CtxGatewayServiceOption(&s.ctx), service.LogGatewayServiceOption(&s.logger))
	if err != nil {
		return err
	}
	s.gateway_serv = *gateway_serv
	return nil
}

func (s *ProxyBrouter) initRouteTbl() error {
	logger := s.logger.WithFields(logrus.Fields{
		"class":  "ProxyRouter",
		"method": "initRouteTbl",
	})
	logger.Infof("size: %d", s.tbl_size)
	var loader RouteTableLoader[manager_model.Proxy] = func(size int) []Route[manager_model.Proxy] {
		s.proxy_serv.Prefetch(s.ctx, size, service.PROXY)
		stream := s.proxy_serv.GetStream()
		proxies := make([]Route[manager_model.Proxy], 0)
		for {
			v := <-stream
			if v == nil {
				break
			}
			route := NewRoute(*v, *NewRouteRule(func(v any) bool { return true }))
			proxies = append(proxies, *route)
		}
		return proxies
	}
	var fb_loader RouteTableLoader[manager_model.Proxy] = func(size int) []Route[manager_model.Proxy] {
		s.proxy_serv.Prefetch(s.ctx, size, service.BACKUP_PROXY)
		stream := s.proxy_serv.GetBackupStream()
		proxies := make([]Route[manager_model.Proxy], 0)
		for {
			v := <-stream
			if v == nil {
				break
			}
			route := NewRoute(*v, *NewRouteRule(func(v any) bool { return true }))
			proxies = append(proxies, *route)
		}
		return proxies
	}

	s.dyn_route_tbl = NewRouteTable(s.tbl_size, s.tbl_cap, LoaderRouteTableOption(loader), LoadFactorRouteTableOption[manager_model.Proxy](0.5), LogRouteTableOption[manager_model.Proxy](&s.logger), NameRouteTableOption[manager_model.Proxy]("proxy"))
	s.dyn_fb_route_tbl = NewRouteTable(s.fb_tbl_size, s.fb_tbl_cap, LoaderRouteTableOption(fb_loader), LoadFactorRouteTableOption[manager_model.Proxy](0.5), LogRouteTableOption[manager_model.Proxy](&s.logger), NameRouteTableOption[manager_model.Proxy]("backup_proxy"))
	return nil
}

func (s *ProxyBrouter) handleCb(cb RouteCallback, opts ...RouteOption) error {
	options := &RouteOptions{}
	for _, opt := range opts {
		opt(options)
	}
	start := time.Now()
	p := s.dyn_route_tbl.Route(opts...)
	if p == nil {
		return ErrNoRoute
	}
	err := cb(p)
	if err != nil {
		err, ok := err.(RouteError)
		if ok {
			s.gateway_serv.CreateEvent(service.BlockedEvent{Time: time.Now(), Proxy: err.from, Site: err.to, Cost: time.Since(start)})
			s.logger.Errorf("failed to callback (err: %+v)", err)
		}
		return err
	} else {
		var src string = ""
		if options.metadata != nil {
			src = (*options.metadata)["addr"]
		}
		s.gateway_serv.CreateEvent(service.PassedEvent{Time: time.Now(), Proxy: p.Ip, Site: src, Cost: time.Since(start)})
		return nil
	}

}

func (s *ProxyBrouter) handlFb(fb RouteCallback, opts ...RouteOption) error {
	options := &RouteOptions{}
	for _, opt := range opts {
		opt(options)
	}
	start := time.Now()
	route := s.dyn_fb_route_tbl.Route(opts...)
	if route == nil {
		return ErrNoRoute
	}
	err := fb(route)
	if err != nil {
		err, ok := err.(RouteError)
		if ok {
			s.gateway_serv.CreateEvent(service.BlockedEvent{Time: time.Now(), Proxy: err.from, Site: err.to, Cost: time.Since(start)})
			s.logger.Errorf("failed to callback (err: %+v)", err)
		}
		return err
	} else {
		var src string = ""
		if options.metadata != nil {
			src = (*options.metadata)["addr"]
		}
		s.gateway_serv.CreateEvent(service.PassedEvent{Time: time.Now(), Proxy: route.Ip, Site: src, Cost: time.Since(start)})
		return nil
	}
}

func (s *ProxyBrouter) Route(ctx context.Context, callback RouteCallback, opts ...RouteOption) {
	options := &RouteOptions{}

	for _, opt := range opts {
		opt(options)
	}
	var max_retry int = default_max_retry
	if options.max_retry != nil {
		max_retry = *options.max_retry
	}
	if s.dyn_route_tbl.Size() > 0 && s.dyn_route_tbl.Size() < max_retry {
		max_retry = s.dyn_route_tbl.Size()
	}
	for i := 0; i < max_retry; i++ {
		err := s.handleCb(callback, opts...)
		//stop proxy routing after proxy routed successfully
		if err == nil {
			return
		}
		//inavaliable proxy incurred failure route, continue to next route
		if errors.As(err, &RouteError{}) {
			continue
		}
		//no route available or other error, fallback to backup proxy
		if options.fallback != nil {
			err = s.handlFb(*options.fallback, opts...)
			if err == nil {
				return
			}
		}
		err = callback(nil)
		if err == nil {
			return
		}
	}
}
