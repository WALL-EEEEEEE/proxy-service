package endpoint

import (
	"context"

	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"

	"github.com/WALL-EEEEEEE/proxy-service/manager/param"
	"github.com/WALL-EEEEEEE/proxy-service/manager/service"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints struct holds the list of endpoints definition
type ProxyProviderServiceEndpoint struct {
	GetProvider    endpoint.Endpoint
	AddProvider    endpoint.Endpoint
	ListProvider   endpoint.Endpoint
	UpdateProvider endpoint.Endpoint
	DeleteProvider endpoint.Endpoint
}

// MakeEndpoints func initializes the Endpoint instances
func NewProxyProviderServiceEndpoint(s service.IProxyProviderService) ProxyProviderServiceEndpoint {
	return ProxyProviderServiceEndpoint{
		GetProvider:    newProxyProviderServiceGetProviderEndpoint(s),
		AddProvider:    newProxyProviderServiceAddProviderEndpoint(s),
		ListProvider:   newProxyProviderServiceListProviderEndpoint(s),
		UpdateProvider: newProxyProviderServiceUpdateProviderEndpoint(s),
		DeleteProvider: newProxyProviderServiceDeleteProviderEndpoint(s),
	}
}

func newProxyProviderServiceGetProviderEndpoint(s service.IProxyProviderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.GetProviderRequest)
		provider, err := s.GetProvider(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		resp := param.GetProviderResponse{}
		resp.StatusResponse = common_param.STATUS_OK
		resp.Provider = *provider
		return resp, nil
	}
}

func newProxyProviderServiceAddProviderEndpoint(s service.IProxyProviderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(param.AddProviderRequest)
		id, err := s.AddProvider(ctx, req.Name)
		if err != nil {
			return nil, err
		}
		resp := param.AddProviderResponse{}
		resp.StatusResponse = common_param.STATUS_OK
		resp.Id = *id
		return resp, nil
	}
}

func newProxyProviderServiceUpdateProviderEndpoint(s service.IProxyProviderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return
	}
}

func newProxyProviderServiceDeleteProviderEndpoint(s service.IProxyProviderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return
	}
}

func newProxyProviderServiceListProviderEndpoint(s service.IProxyProviderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(param.ListProviderRequest)
		provider_list, err := s.ListProvider(ctx, req.Limit, req.Offset)
		if err != nil {
			return nil, err
		}
		resp := param.ListProviderResponse{}
		resp.StatusResponse = common_param.STATUS_OK
		resp.ProviderList = provider_list
		return resp, nil
	}
}
