package endpoint

import (
	"context"

	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"

	"github.com/WALL-EEEEEEE/proxy-service/provider-adapter/param"
	"github.com/WALL-EEEEEEE/proxy-service/provider-adapter/service"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints struct holds the list of endpoints definition
type AdapterServiceEndpoint struct {
	ListProxies endpoint.Endpoint
}

// MakeEndpoints func initializes the Endpoint instances
func NewAdapterServiceEndpoint(s service.IProxyAdapterService) AdapterServiceEndpoint {
	return AdapterServiceEndpoint{
		ListProxies: newAdapterServiceListProxiesEndpoint(s),
	}
}

func newAdapterServiceListProxiesEndpoint(s service.IProxyAdapterService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.ListProxiesRequest)
		total, offset, proxies, err := s.ListProxies(ctx, req.Offset, req.Limit, req.Params)
		if err != nil {
			return nil, err
		}
		resp := param.ListProxiesResponse{Status: common_param.STATUS_OK}
		resp.Page.Total = total
		resp.Page.Limit = req.Limit
		resp.Page.Offset = offset
		resp.Proxies = proxies
		return resp, nil
	}
}
