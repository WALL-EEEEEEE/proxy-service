package transport

import (
	"context"

	ends "github.com/WALL-EEEEEEE/proxy-service/manager/endpoint"
	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/WALL-EEEEEEE/proxy-service/manager/param"

	gt "github.com/go-kit/kit/transport/grpc"
	"github.com/sirupsen/logrus"
)

type ProxyApiServiceEndpoint = ends.ProxyApiServiceEndpoint

type ProxyApiServiceTransport struct {
	get_api             gt.Handler
	get_api_by_provider gt.Handler
	add_api             gt.Handler
	delete_api          gt.Handler
	update_api          gt.Handler
	pb.UnimplementedProxyApiServiceServer
}

// NewProxyQueryTransport initializes a new ProxyQuery Transport
func NewProxyApiServiceTransport(endpoint ProxyApiServiceEndpoint, logger *logrus.Logger) pb.ProxyApiServiceServer {
	return &ProxyApiServiceTransport{
		get_api: gt.NewServer(
			endpoint.GetApi,
			decodeProxyApiServiceGetApiRequest,
			encodeProxyApiServiceGetApiResponse,
		),
		get_api_by_provider: gt.NewServer(
			endpoint.GetApiByProvider,
			decodeProxyApiServiceGetApiByProviderRequest,
			encodeProxyApiServiceGetApiByProviderResponse,
		),
		add_api: gt.NewServer(
			endpoint.AddApi,
			decodeProxyApiServiceAddApiRequest,
			encodeProxyApiServiceAddApiResponse,
		),
	}
}

func (s *ProxyApiServiceTransport) GetApi(ctx context.Context, req *pb.GetApiRequest) (*pb.GetApiResponse, error) {
	_, resp, err := s.get_api.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.GetApiResponse), nil
}

func (s *ProxyApiServiceTransport) GetApiByProvider(ctx context.Context, req *pb.GetApiByProviderRequest) (*pb.GetApiByProviderResponse, error) {
	_, resp, err := s.get_api_by_provider.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.GetApiByProviderResponse), nil
}

func decodeProxyApiServiceGetApiByProviderRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.GetApiByProviderRequest)
	var get_by_provider_req param.GetApiByProviderRequest = param.GetApiByProviderRequest{
		ProviderId: req.ProviderId,
	}
	return get_by_provider_req, nil
}

func encodeProxyApiServiceGetApiByProviderResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.GetApiByProviderResponse)
	ret_resp := &pb.GetApiByProviderResponse{Status: &pb.ResponseStatus{Code: resp.Code, Result: resp.Result, Message: resp.Message}}

	if resp.Apis == nil {
		return ret_resp, nil
	}
	var ret_apis []*pb.ProxyAPI
	for _, api := range resp.Apis {
		ret_api := &pb.ProxyAPI{
			Name:     api.Name,
			Interval: api.UpdateInterval,
			Service: &pb.Service{
				Host:   api.Service.Host,
				Name:   api.Service.Name,
				Params: api.Service.Params,
			},
		}
		ret_apis = append(ret_apis, ret_api)
	}
	ret_resp.ProxyApis = ret_apis
	return ret_resp, nil
}

func decodeProxyApiServiceGetApiRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.GetApiRequest)
	var get_req param.GetApiRequest = param.GetApiRequest{
		Id: req.Id,
	}
	return get_req, nil
}

func encodeProxyApiServiceGetApiResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.GetApiResponse)
	ret_resp := &pb.GetApiResponse{Status: &pb.ResponseStatus{Code: resp.Code, Result: resp.Result, Message: resp.Message}}
	ret_resp.Api = &pb.ProxyAPI{
		Name:     resp.Api.Name,
		Interval: resp.Api.UpdateInterval,
		Service: &pb.Service{
			Host:   resp.Api.Service.Host,
			Name:   resp.Api.Service.Name,
			Params: resp.Api.Service.Params,
		},
	}

	return ret_resp, nil
}

func (s *ProxyApiServiceTransport) AddApi(ctx context.Context, req *pb.AddApiRequest) (*pb.AddApiResponse, error) {
	_, resp, err := s.add_api.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.AddApiResponse), nil
}

func decodeProxyApiServiceAddApiRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.AddApiRequest)
	var add_req param.AddApiRequest = param.AddApiRequest{
		Name:           req.Api.Name,
		ProviderId:     req.ProviderId,
		UpdateInterval: req.Api.Interval,
		Service: model.Service{
			Host:   req.Api.Service.Host,
			Name:   req.Api.Service.Name,
			Params: req.Api.Service.Params,
		},
	}
	return add_req, nil
}

func encodeProxyApiServiceAddApiResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.AddApiResponse)
	return &pb.AddApiResponse{Status: &pb.ResponseStatus{Code: resp.Code, Result: resp.Result, Message: resp.Message}, Id: resp.Id}, nil
}
