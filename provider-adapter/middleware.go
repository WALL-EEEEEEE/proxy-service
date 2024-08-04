package main

import (
	"context"
	"fmt"

	"github.com/WALL-EEEEEEE/proxy-service/provider-adapter/param"
	servs "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-kit/kit/endpoint"
)

func RoutingMiddleware(serv servs.IProxyAdapterService) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			req := request.(param.ListProxiesRequest)
			if req.Adapter != serv.GetName() {
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("adapter %s not found", req.Adapter))
			}
			return next(ctx, request)
		}
	}
}

/*
type LoggingServiceMiddleware struct {
	logger log.Logger
	next   servs.IProxyAdapterService
}

func (mw LoggingServiceMiddleware) ListProxies(ctx context.Context, offset, limit int64, params map[string]string) (ret_total int64, ret_offset int64, proxies []model.Proxy, err error) {
	defer func(begin time.Time) {
		logger := log.With(mw.logger,
			"adapter", mw.next.GetName(),
			"method", "ListProxies",
			"limit", limit,
			"offset", offset,
			"params", params,
			"took", time.Since(begin))
		if err == nil {
			logger.Log("msg", fmt.Sprintf("got %d proxies", len(proxies)))
		} else {
			level.Error(logger).Log("msg", "failed to get proxies", "err", err)
		}
	}(time.Now())
	ret_total, ret_offset, proxies, err = mw.next.ListProxies(ctx, offset, limit, params)
	return
}
func (mw LoggingServiceMiddleware) GetName() string {
	return mw.next.GetName()
}

*/
