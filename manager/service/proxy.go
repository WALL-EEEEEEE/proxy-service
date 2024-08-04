package service

// Proxy provides operations on proxy.
import (
	"context"
	"errors"
	"fmt"

	"github.com/WALL-EEEEEEE/proxy-service/common"

	"github.com/WALL-EEEEEEE/proxy-service/manager/cache"
	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	eris_format = eris.NewDefaultStringFormat(eris.FormatOptions{
		InvertOutput: true, // flag that inverts the error output (wrap errors shown first)
		WithTrace:    true, // flag that enables stack trace output
		InvertTrace:  true, // flag that inverts the stack trace output (top of call stack shown first)
	})
)

type Filter = common.Filter[any]

type Paginator common.Paginator[model.Proxy]

type IProxyService interface {
	ListProxies(context.Context, int64, int64, *pb.Filter) (*Paginator, error)
	GetProxy(context.Context, string) (*model.Proxy, error)
	GetProxyByIp(context.Context, string) (*model.Proxy, error)
	AddProxy(context.Context, model.Proxy) (*string, error)
	UpdateProxy(context.Context, string, model.Proxy, []string) error
	DeleteProxy(context.Context, string) error
}

type ProxyService struct {
	logger      *logrus.Logger
	proxy_store *cache.ProxyStore
}

func NewProxyService(logger *logrus.Logger, proxy_store *cache.ProxyStore) *ProxyService {
	return &ProxyService{
		logger:      logger,
		proxy_store: proxy_store,
	}
}

func (p ProxyService) ListProxies(ctx context.Context, limit, offset int64, filter *pb.Filter) (*Paginator, error) {
	_ = logrus.WithFields(logrus.Fields{
		"class":  "ProxyService",
		"method": "ListProxies",
	})
	var proxies []model.Proxy
	var paginator = Paginator{
		Total:  0,
		Limit:  limit,
		Offset: offset,
		Items:  proxies,
	}
	_paginator := common.Paginator[model.Proxy](paginator)
	err := p.proxy_store.ListWithFilters(ctx, &_paginator, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed list proxy with filters %+v (err: %s)", filter, err.Error()))
	}
	ret_paginator := Paginator(_paginator)
	return &ret_paginator, nil
}

func (p ProxyService) GetProxy(ctx context.Context, id string) (*model.Proxy, error) {
	proxy := model.Proxy{}
	err := p.proxy_store.GetById(ctx, id, &proxy)
	if err != nil {
		if errors.Is(err, cache.ProxyNotFoundError) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("proxy with id %s doesn't exists", id))
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed get proxy by id %s (err: %s)", id, err.Error()))
	}
	return &proxy, nil
}

func (p ProxyService) GetProxyByIp(ctx context.Context, ip string) (*model.Proxy, error) {
	proxy := model.Proxy{}
	err := p.proxy_store.GetByIp(ctx, ip, &proxy)
	if err != nil {
		if errors.Is(err, cache.ProxyNotFoundError) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("proxy with ip %s doesn't exists", ip))
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed get proxy %s (err: %s)", ip, err.Error()))
	}
	return &proxy, nil
}

func (p ProxyService) AddProxy(ctx context.Context, proxy model.Proxy) (*string, error) {
	logger := logrus.WithFields(logrus.Fields{
		"class":  "ProxyService",
		"method": "AddProxy",
	})
	if_exists, err := p.proxy_store.ExistsIp(ctx, proxy.Ip)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add proxy %s (err: %s)", proxy.Ip, err.Error()))
	}
	if if_exists {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("proxy %s already exists", proxy.Ip))
	}
	proxy_id, err := p.proxy_store.Add(ctx, &proxy)
	if err != nil {
		logger.Errorf("%+v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add proxy %s (err: %s)", proxy.Ip, err.Error()))
	}
	return proxy_id, nil
}

func (p ProxyService) UpdateProxy(ctx context.Context, id string, proxy model.Proxy, paths []string) error {
	logger := logrus.WithFields(logrus.Fields{
		"class":  "ProxyService",
		"method": "UpdateProxy",
		"paths":  paths,
	})
	logger.Infof("%#v", proxy)
	err := p.proxy_store.Update(ctx, id, proxy, paths)
	if err != nil {
		if errors.Is(err, cache.ProxyNotFoundError) {
			return status.Error(codes.NotFound, fmt.Sprintf("proxy with id %s doesn't exists", id))
		}
		return status.Error(codes.Internal, fmt.Sprintf("failed to update proxy %s (error: %s)", id, err.Error()))
	}
	return nil
}
func (p ProxyService) DeleteProxy(ctx context.Context, id string) error {
	return nil
}
