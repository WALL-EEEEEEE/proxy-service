package transport

import (
	"context"

	ends "github.com/WALL-EEEEEEE/proxy-service/manager/endpoint"
	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/param"

	gt "github.com/go-kit/kit/transport/grpc"
	"github.com/sirupsen/logrus"
)

type ProxyProviderServiceEndpoint = ends.ProxyProviderServiceEndpoint

type ProxyProviderServiceTransport struct {
	get_provider    gt.Handler
	add_provider    gt.Handler
	list_provider   gt.Handler
	delete_provider gt.Handler
	update_provider gt.Handler
	pb.UnimplementedProxyProviderServiceServer
}

// NewProxyQueryTransport initializes a new ProxyQuery Transport
func NewProxyProviderServiceTransport(endpoint ProxyProviderServiceEndpoint, logger *logrus.Logger) pb.ProxyProviderServiceServer {
	return &ProxyProviderServiceTransport{
		get_provider: gt.NewServer(
			endpoint.GetProvider,
			decodeProxyProviderServiceGetProviderRequest,
			encodeProxyProviderServiceGetProviderResponse,
		),
		add_provider: gt.NewServer(
			endpoint.AddProvider,
			decodeProxyProviderServiceAddProviderRequest,
			encodeProxyProviderServiceAddProviderResponse,
		),
		list_provider: gt.NewServer(
			endpoint.ListProvider,
			decodeProxyProviderServiceListProviderRequest,
			encodeProxyProviderServiceListProviderResponse,
		),
	}
}

func (s *ProxyProviderServiceTransport) GetProvider(ctx context.Context, req *pb.GetProviderRequest) (*pb.GetProviderResponse, error) {
	_, resp, err := s.get_provider.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.GetProviderResponse), nil
}

func decodeProxyProviderServiceGetProviderRequest(_ context.Context, request interface{}) (interface{}, error) {
	var get_provider_req param.GetProviderRequest
	req := request.(*pb.GetProviderRequest)
	get_provider_req.Id = req.Id
	return get_provider_req, nil
}

func encodeProxyProviderServiceGetProviderResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.GetProviderResponse)
	ret_resp := pb.GetProviderResponse{Status: &pb.ResponseStatus{Code: resp.Code, Result: resp.Result, Message: resp.Message}}
	ret_resp.Provider = &pb.ProxyProvider{Id: resp.Provider.Id, Name: resp.Provider.Name}
	return &ret_resp, nil
}

func (s *ProxyProviderServiceTransport) AddProvider(ctx context.Context, req *pb.AddProviderRequest) (*pb.AddProviderResponse, error) {
	_, resp, err := s.add_provider.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.AddProviderResponse), nil
}

func decodeProxyProviderServiceAddProviderRequest(_ context.Context, request interface{}) (interface{}, error) {
	var add_req param.AddProviderRequest
	req := request.(*pb.AddProviderRequest)
	add_req.Name = req.Name
	return add_req, nil
}

func encodeProxyProviderServiceAddProviderResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.AddProviderResponse)
	return &pb.AddProviderResponse{Status: &pb.ResponseStatus{Code: resp.Code, Result: resp.Result, Message: resp.Message}, Id: resp.Id}, nil
}

func (s *ProxyProviderServiceTransport) ListProvider(ctx context.Context, req *pb.ListProviderRequest) (*pb.ListProviderResponse, error) {
	_, resp, err := s.list_provider.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ListProviderResponse), nil
}

func decodeProxyProviderServiceListProviderRequest(_ context.Context, request interface{}) (interface{}, error) {
	var list_provider_req param.ListProviderRequest
	req := request.(*pb.ListProviderRequest)
	list_provider_req.Limit = req.Limit
	list_provider_req.Offset = req.Offset
	return list_provider_req, nil
}

func encodeProxyProviderServiceListProviderResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.ListProviderResponse)
	ret_resp := &pb.ListProviderResponse{Status: &pb.ResponseStatus{Code: resp.Code, Result: resp.Result, Message: resp.Message}}
	if resp.ProviderList == nil {
		ret_resp.ProviderList = []*pb.ProxyProvider{}
	}
	for _, provider := range resp.ProviderList {
		ret_resp.ProviderList = append(ret_resp.ProviderList, &pb.ProxyProvider{Id: provider.Id, Name: provider.Name})
	}
	return ret_resp, nil
}
