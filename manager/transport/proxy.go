package transport

import (
	"context"

	"github.com/WALL-EEEEEEE/proxy-service/common"

	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"

	ends "github.com/WALL-EEEEEEE/proxy-service/manager/endpoint"
	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/WALL-EEEEEEE/proxy-service/manager/param"
	"github.com/WALL-EEEEEEE/proxy-service/manager/util"

	"github.com/bobg/go-generics/slices"
	gt "github.com/go-kit/kit/transport/grpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	DEFAULT_LIST_MASK = []string{"id", "proto", "ip", "port", "status", "provider", "api", "provider_id", "api_id", "attr", "created_at", "updated_at", "checked_at", "expire_time", "use_config"}
)

type ProxyServiceTransport struct {
	list_proxies    gt.Handler
	add_proxy       gt.Handler
	delete_proxy    gt.Handler
	update_proxy    gt.Handler
	get_proxy       gt.Handler
	get_proxy_by_ip gt.Handler
	pb.UnimplementedProxyServiceServer
}

// NewProxyQueryTransport initializes a new ProxyQuery Transport
func NewProxyServiceTransport(endpoint ends.ProxyServiceEndpoint, logger *logrus.Logger) pb.ProxyServiceServer {
	return &ProxyServiceTransport{
		list_proxies: gt.NewServer(
			endpoint.ListProxies,
			decodeProxyServiceListProxiesRequest,
			encodeProxyServiceListProxiesResponse,
		),
		add_proxy: gt.NewServer(
			endpoint.AddProxy,
			decodeProxyServiceAddProxyRequest,
			encodeProxyServiceAddProxyResponse,
		),
		update_proxy: gt.NewServer(
			endpoint.UpdateProxy,
			decodeProxyServiceUpdateProxyRequest,
			encodeProxyServiceUpdateProxyResponse,
		),
		get_proxy: gt.NewServer(
			endpoint.GetProxy,
			decodeProxyServiceGetProxyRequest,
			encodeProxyServiceGetProxyResponse,
		),
		get_proxy_by_ip: gt.NewServer(
			endpoint.GetProxyByIp,
			decodeProxyServiceGetProxyByIpRequest,
			encodeProxyServiceGetProxyByIpResponse,
		),
	}
}

func (s *ProxyServiceTransport) ListProxies(ctx context.Context, req *pb.ListProxiesRequest) (*pb.ListProxiesResponse, error) {
	_, resp, err := s.list_proxies.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ListProxiesResponse), nil
}

func (s *ProxyServiceTransport) AddProxy(ctx context.Context, req *pb.AddProxyRequest) (*pb.AddProxyResponse, error) {
	_, resp, err := s.add_proxy.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.AddProxyResponse), nil
}

func (s *ProxyServiceTransport) UpdateProxy(ctx context.Context, req *pb.UpdateProxyRequest) (*pb.UpdateProxyResponse, error) {
	_, resp, err := s.update_proxy.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.UpdateProxyResponse), nil
}

func (s *ProxyServiceTransport) GetProxy(ctx context.Context, req *pb.GetProxyRequest) (*pb.GetProxyResponse, error) {
	_, resp, err := s.get_proxy.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.GetProxyResponse), nil
}

func (s *ProxyServiceTransport) GetProxyByIp(ctx context.Context, req *pb.GetProxyByIpRequest) (*pb.GetProxyByIpResponse, error) {
	_, resp, err := s.get_proxy_by_ip.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.GetProxyByIpResponse), nil
}

func decodeProxyServiceListProxiesRequest(_ context.Context, request interface{}) (interface{}, error) {
	var err error
	req := request.(*pb.ListProxiesRequest)
	var list_req param.ListProxiesRequest = param.ListProxiesRequest{}
	if req.Query != nil {
		list_req.Pager = common_param.Pager{Limit: req.Query.Limit, Offset: int64(req.Query.Offset)}
		list_req.Filter = req.Query.Filter

	}
	if req.Fields != nil {
		list_req.ListMask = req.Fields.Paths
	} else {
		list_req.ListMask = DEFAULT_LIST_MASK
	}
	return list_req, err
}

func encodeProxyServiceListProxiesResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.ListProxiesResponse)
	proxies, err := slices.Map[model.Proxy, *pb.Proxy](resp.ProxyList, func(i int, p model.Proxy) (*pb.Proxy, error) {
		var err error = nil
		proxy := util.PbFromProxy(&p)
		if resp.ListMask != nil && len(resp.ListMask) > 0 {
			proxy, err = common.MaskFields(proxy, resp.ListMask)
		}
		return proxy, err
	})
	return &pb.ListProxiesResponse{Status: &pb.ResponseStatus{Result: resp.Result, Message: resp.Message, Code: resp.Code}, Limit: resp.Limit, Offset: resp.Offset, Count: resp.Count, Total: resp.Total, ProxyList: proxies}, err
}

func decodeProxyServiceAddProxyRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.AddProxyRequest)
	ret_req := util.AddProxyRequestFromPb(req)
	if ret_req == nil {
		return nil, nil
	}
	return *ret_req, nil
}

func encodeProxyServiceAddProxyResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.AddProxyResponse)

	return &pb.AddProxyResponse{Status: &pb.ResponseStatus{Result: resp.Result, Message: resp.Message, Code: resp.Code}, Id: resp.Id}, nil
}

func decodeProxyServiceUpdateProxyRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.UpdateProxyRequest)
	logrus.Debugf("%+v", req.Fields)
	mask_proxy, err := common.MaskFields(req.Proxy, req.Fields.GetPaths())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	req.Proxy = mask_proxy
	ret_req := util.UpdateProxyRequestFromPb(req)
	if ret_req == nil {
		return nil, nil
	}
	return *ret_req, nil
}

func encodeProxyServiceUpdateProxyResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.UpdateProxyResponse)
	return &pb.UpdateProxyResponse{Status: &pb.ResponseStatus{Result: resp.Result, Message: resp.Message, Code: resp.Code}}, nil
}

func decodeProxyServiceGetProxyRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.GetProxyRequest)

	var proxy_query_req param.GetProxyRequest = param.GetProxyRequest{
		Id: req.Id,
	}
	return proxy_query_req, nil
}

func encodeProxyServiceGetProxyResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.GetProxyResponse)
	proxy := util.PbFromProxy(&resp.Proxy)
	return &pb.GetProxyResponse{Status: &pb.ResponseStatus{Result: resp.Result, Message: resp.Message, Code: resp.Code}, Proxy: proxy}, nil
}

func decodeProxyServiceGetProxyByIpRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.GetProxyByIpRequest)
	var proxy_get_req param.GetProxyByIpRequest = param.GetProxyByIpRequest{
		Ip: req.Ip,
	}
	return proxy_get_req, nil
}

func encodeProxyServiceGetProxyByIpResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.GetProxyByIpResponse)
	proxy := util.PbFromProxy(&resp.Proxy)
	return &pb.GetProxyByIpResponse{Status: &pb.ResponseStatus{Result: resp.Result, Message: resp.Message, Code: resp.Code}, Proxy: proxy}, nil
}
