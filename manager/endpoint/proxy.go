package endpoint

import (
	"context"

	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/WALL-EEEEEEE/proxy-service/manager/param"
	"github.com/WALL-EEEEEEE/proxy-service/manager/service"

	"github.com/go-kit/kit/endpoint"
)

type Proxy = model.Proxy

// Endpoints struct holds the list of endpoints definition
type ProxyServiceEndpoint struct {
	ListProxies  endpoint.Endpoint
	AddProxy     endpoint.Endpoint
	UpdateProxy  endpoint.Endpoint
	DeleteProxy  endpoint.Endpoint
	GetProxy     endpoint.Endpoint
	GetProxyByIp endpoint.Endpoint
}

// MakeEndpoints func initializes the Endpoint instances
func NewProxyServiceEndpoint(s service.IProxyService) ProxyServiceEndpoint {
	return ProxyServiceEndpoint{
		ListProxies:  newProxyServiceListProxiesEndpoint(s),
		AddProxy:     newProxyServiceAddProxyEndpoint(s),
		UpdateProxy:  newProxyServiceUpdateProxyEndpoint(s),
		DeleteProxy:  newProxyServiceDeleteProxyEndpoint(s),
		GetProxy:     newProxyServiceGetProxyEndpoint(s),
		GetProxyByIp: newProxyServiceGetProxyByIpEndpoint(s),
	}
}

func newProxyServiceListProxiesEndpoint(s service.IProxyService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.ListProxiesRequest)
		paginator, err := s.ListProxies(ctx, req.Limit, req.Offset, req.Filter)
		if err != nil {
			return nil, err

		}
		resp := param.ListProxiesResponse{}
		resp.ListMask = req.ListMask
		resp.StatusResponse = common_param.STATUS_OK
		resp.ProxyList = paginator.Items
		resp.Limit = paginator.Limit
		resp.Offset = paginator.Offset
		resp.Total = paginator.Total
		resp.Count = paginator.Count
		response = resp
		return
	}
}

func newProxyServiceAddProxyEndpoint(s service.IProxyService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.AddProxyRequest)
		id, err := s.AddProxy(ctx, req.Proxy)
		if err != nil {
			return nil, err
		}
		resp := param.AddProxyResponse{}
		resp.StatusResponse = common_param.STATUS_OK
		resp.Id = *id
		response = resp
		return
	}
}

func newProxyServiceUpdateProxyEndpoint(s service.IProxyService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.UpdateProxyRequest)
		err = s.UpdateProxy(ctx, req.Id, req.Proxy, req.UpdateMask)
		if err != nil {
			return nil, err
		}
		resp := param.UpdateProxyResponse{}
		resp.StatusResponse = common_param.STATUS_OK
		response = resp
		return
	}
}

func newProxyServiceGetProxyEndpoint(s service.IProxyService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.GetProxyRequest)
		proxy, err := s.GetProxy(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		resp := param.GetProxyResponse{}
		resp.StatusResponse = common_param.STATUS_OK
		resp.Proxy = *proxy
		response = resp
		return
	}
}

func newProxyServiceGetProxyByIpEndpoint(s service.IProxyService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.GetProxyByIpRequest)
		proxy, err := s.GetProxyByIp(ctx, req.Ip)
		if err != nil {
			return nil, err
		}
		resp := param.GetProxyByIpResponse{}
		resp.StatusResponse = common_param.STATUS_OK
		resp.Proxy = *proxy
		response = resp
		return
	}
}

func newProxyServiceDeleteProxyEndpoint(s service.IProxyService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return
	}
}
