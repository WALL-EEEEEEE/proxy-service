package service

import (
	"context"
	"errors"
	"time"

	managerv1 "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	manager_model "github.com/WALL-EEEEEEE/proxy-service/manager/model"
	manager_util "github.com/WALL-EEEEEEE/proxy-service/manager/util"
	"github.com/sirupsen/logrus"

	"github.com/WALL-EEEEEEE/proxy-service/gateway/internal/client"
	"github.com/WALL-EEEEEEE/proxy-service/gateway/log"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
)

const (
	default_size              = 20
	default_prefetch_size     = 20
	default_prefetch_interval = time.Duration(3) * time.Second
)

type prefetchMode int

const (
	PROXY prefetchMode = iota + 1
	BACKUP_PROXY
)

var ErrInvalidPrefetchMode = errors.New("invalid prefetch mode")

type ProxyServiceOptions struct {
	logger            *log.Logger
	ctx               *context.Context
	prefetch_interval *time.Duration
	size              *int
}

type ProxyServiceOption func(*ProxyServiceOptions)

func LogProxyServiceOption(logger *log.Logger) ProxyServiceOption {
	return func(options *ProxyServiceOptions) {
		options.logger = logger
	}
}
func CtxProxyServiceOption(ctx *context.Context) ProxyServiceOption {
	return func(options *ProxyServiceOptions) {
		options.ctx = ctx
	}
}

func SizeProxyServiceOption(size int) ProxyServiceOption {
	return func(options *ProxyServiceOptions) {
		options.size = &size
	}
}

func PrefetchIntervalProxyServiceOption(interval time.Duration) ProxyServiceOption {
	return func(options *ProxyServiceOptions) {
		options.prefetch_interval = &interval
	}
}

type ProxyService struct {
	ctx                  context.Context
	client               client.ProxyClient
	logger               log.Logger
	pos                  int
	size                 int
	prefetch_interval    time.Duration
	prefetch_chan        chan int
	prefetch_backup_chan chan int
	stream               chan *manager_model.Proxy
	backup_stream        chan *manager_model.Proxy
}

func NewProxyService(grpc_addr string, opts ...ProxyServiceOption) (*ProxyService, error) {
	options := &ProxyServiceOptions{}
	for _, opt := range opts {
		opt(options)
	}
	client, err := client.NewProxyClient(grpc_addr, client.CtxProxyClientOption(options.ctx), client.LogProxyClientOption(options.logger))
	if err != nil {
		return nil, err
	}
	prefetch_chan := make(chan int)
	prefetch_backup_chan := make(chan int)
	stream := make(chan *manager_model.Proxy)
	backup_stream := make(chan *manager_model.Proxy)
	service := &ProxyService{client: *client, pos: 0, prefetch_chan: prefetch_chan, prefetch_backup_chan: prefetch_backup_chan, stream: stream, backup_stream: backup_stream}
	if options.logger != nil {
		service.logger = *options.logger
	} else {
		service.logger = log.DefaultLogger
	}
	if options.prefetch_interval != nil {
		service.prefetch_interval = *options.prefetch_interval
	} else {
		service.prefetch_interval = default_prefetch_interval
	}
	if options.ctx != nil {
		service.ctx = *options.ctx
	} else {
		service.ctx = context.Background()
	}
	if options.size != nil {
		service.size = *options.size
	} else {
		service.size = default_size
	}
	service.initPrefetcher(service.ctx)
	return service, nil
}

func (s *ProxyService) initPrefetcher(ctx context.Context) {
	logger := s.logger.WithFields(logrus.Fields{
		"class":  "ProxyService",
		"method": "Prefetch",
	})
	prefetch_func := func(size int) {
		logger := logger.WithFields(logrus.Fields{
			"mode": "Proxy",
		})
		if size <= 0 {
			logger.Errorf("invalid prefetch size: %d (prefetch size must greater than 0)", size)
			return
		}
		offset := s.pos
		logger.WithFields(
			logrus.Fields{
				"offset": offset,
				"size":   size,
			}).Info()
		proxies, err := s.ListProxies(ctx, size, offset)
		if err != nil {
			status, _ := grpc_status.FromError(err)
			if status.Code() == codes.Unavailable {
				logger.Warnf("%s service is unavailable", s.client.GetAddr())
			} else {
				logger.Warnf("%s service error (error: %+v)", s.client.GetAddr(), err)
			}
			return
		}
		if proxies == nil {
			logger.Warnf("no proxies found from service %s", s.client.GetAddr())
			return
		}
		for i := range proxies {
			//@Fix: use index to copy items from proxies to stream, cause for-loop variables are reused )
			s.stream <- &proxies[i]
			s.pos++
		}
		s.stream <- nil
		logger.WithFields(
			logrus.Fields{
				"offset": s.pos,
				"size":   size,
			}).Infof("prefetched: %d", s.pos-offset)
	}
	prefetch_bachup_func := func(size int) {
		logger := logger.WithFields(logrus.Fields{
			"mode": "BackupProxy",
		})
		if size <= 0 {
			logger.Errorf("invalid prefetch size: %d (prefetch size must greater than 0)", size)
			return
		}
		offset := s.pos
		logger.WithFields(
			logrus.Fields{
				"offset": offset,
				"size":   size,
			}).Info()
		proxies, err := s.ListBackupProxies(ctx, size, offset)
		if err != nil {
			status, _ := grpc_status.FromError(err)
			if status.Code() == codes.Unavailable {
				logger.Warnf("%s service is unavailable", s.client.GetAddr())
			} else {
				logger.Warnf("%s service error (error: %+v)", s.client.GetAddr(), err)
			}
			return
		}
		if proxies == nil {
			logger.Warnf("no proxies found from service %s", s.client.GetAddr())
			return
		}
		for i := range proxies {
			//@Fix: use index to copy items from proxies to stream, cause for-loop variables are reused )
			s.backup_stream <- &proxies[i]
			s.pos++
		}
		s.backup_stream <- nil
		logger.WithFields(
			logrus.Fields{
				"offset": s.pos,
				"size":   size,
			}).Infof("prefetched: %d", s.pos-offset)
	}

	go func() {
		breakLoop := false
		for {
			select {
			case size := <-s.prefetch_chan:
				prefetch_func(size)
			case bsize := <-s.prefetch_backup_chan:
				prefetch_bachup_func(bsize)
			case <-ctx.Done():
				breakLoop = true
			}
			if breakLoop {
				break
			}
		}
	}()
}

func (s *ProxyService) Prefetch(ctx context.Context, size int, mode prefetchMode) error {
	switch mode {
	case PROXY:
		s.prefetch_chan <- size
	case BACKUP_PROXY:
		s.prefetch_backup_chan <- size
	default:
		return ErrInvalidPrefetchMode
	}
	return nil
}

func (s *ProxyService) GetStream() <-chan *manager_model.Proxy {
	return s.stream
}

func (s *ProxyService) GetBackupStream() <-chan *manager_model.Proxy {
	return s.backup_stream
}

func (s *ProxyService) ListProxies(ctx context.Context, limit int, offset int) ([]manager_model.Proxy, error) {
	var ret []manager_model.Proxy
	filter_options := client.FilterListProxiesOption(
		client.ConstructPropertyFilter("tags", managerv1.PropertyFilter_EQUAL, "ip"),
		client.ConstructPropertyFilter("proto", managerv1.PropertyFilter_EQUAL, managerv1.Proto_PROTO_HTTP.String()),
	)
	proxies, err := s.client.ListProxies(ctx, limit, offset, filter_options)
	for _, p := range proxies {
		ret = append(ret, *manager_util.ProxyFromPb(p))
	}
	return ret, err
}

func (s *ProxyService) ListBackupProxies(ctx context.Context, limit int, offset int) ([]manager_model.Proxy, error) {
	var ret []manager_model.Proxy
	filter_options := client.FilterListProxiesOption(
		client.ConstructPropertyFilter("tags", managerv1.PropertyFilter_EQUAL, "gateway"),
		client.ConstructPropertyFilter("proto", managerv1.PropertyFilter_EQUAL, managerv1.Proto_PROTO_HTTP.String()),
	)
	proxies, err := s.client.ListProxies(ctx, limit, offset, filter_options)
	for _, p := range proxies {
		ret = append(ret, *manager_util.ProxyFromPb(p))
	}
	return ret, err
}
