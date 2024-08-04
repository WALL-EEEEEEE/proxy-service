package endpoint

import (
	"context"

	. "github.com/WALL-EEEEEEE/proxy-service/common/param"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/WALL-EEEEEEE/proxy-service/manager/param"
	"github.com/WALL-EEEEEEE/proxy-service/manager/service"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints struct holds the list of endpoints definition
type ProxyApiServiceEndpoint struct {
	GetApi           endpoint.Endpoint
	GetApiByProvider endpoint.Endpoint
	AddApi           endpoint.Endpoint
	UpdateApi        endpoint.Endpoint
	DeleteApi        endpoint.Endpoint
}

// MakeEndpoints func initializes the Endpoint instances
func NewProxyApiServiceEndpoint(s service.IProxyApiService) ProxyApiServiceEndpoint {
	return ProxyApiServiceEndpoint{
		GetApi:           newProxyApiServiceGetApiEndpoint(s),
		GetApiByProvider: newProxyApiServiceGetApiByProviderEndpoint(s),
		AddApi:           newProxyApiServiceAddApiEndpoint(s),
		UpdateApi:        newProxyApiServiceUpdateApiEndpoint(s),
		DeleteApi:        newProxyApiServiceDeleteApiEndpoint(s),
	}
}

func newProxyApiServiceGetApiEndpoint(s service.IProxyApiService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.GetApiRequest)
		api, err := s.GetApi(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		resp := param.GetApiResponse{}
		resp.StatusResponse = STATUS_OK
		resp.Api = *api
		response = resp
		return
	}
}

func newProxyApiServiceGetApiByProviderEndpoint(s service.IProxyApiService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.GetApiByProviderRequest)
		apis, err := s.GetApiByProvider(ctx, req.ProviderId)
		if err != nil {
			return nil, err
		}
		resp := param.GetApiByProviderResponse{}
		resp.StatusResponse = STATUS_OK
		resp.Apis = apis
		response = resp
		return
	}
}

func newProxyApiServiceAddApiEndpoint(s service.IProxyApiService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(param.AddApiRequest)

		proxy_api := model.ProxyApi{
			ProviderId: req.ProviderId,
			Name:       req.Name,
			Service: model.Service{
				Host:   req.Service.Host,
				Params: req.Service.Params,
				Name:   req.Service.Name,
			},
			UpdateInterval: req.UpdateInterval,
		}
		id, err := s.AddApi(ctx, proxy_api)
		if err != nil {
			return nil, err
		}
		resp := param.AddApiResponse{}
		resp.StatusResponse = STATUS_OK
		resp.Id = *id
		return resp, nil
	}
}

func newProxyApiServiceUpdateApiEndpoint(s service.IProxyApiService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return
	}
}

func newProxyApiServiceDeleteApiEndpoint(s service.IProxyApiService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return
	}
}
