package transport

import (
	"context"

	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"

	ends "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/endpoint"
	pb "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/gen/adapter/v1"
	"github.com/WALL-EEEEEEE/proxy-service/provider-adapter/param"

	manager_pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"

	"github.com/bobg/go-generics/slices"
	gt "github.com/go-kit/kit/transport/grpc"
	log "github.com/sirupsen/logrus"
)

type AdapterServiceEndpoint = ends.AdapterServiceEndpoint

type AdapterServiceTransport struct {
	list_proxies gt.Handler
	pb.UnimplementedAdapterServiceServer
}

// NewProxyQueryTransport initializes a new ProxyQuery Transport
func NewAdapterServiceTransport(logger *log.Logger, endpoint AdapterServiceEndpoint) pb.AdapterServiceServer {
	return &AdapterServiceTransport{
		list_proxies: gt.NewServer(
			endpoint.ListProxies,
			decodeAdapterServiceListProxiesRequest,
			encodeAdapterServiceListProxiesResponse,
		),
	}
}

func (t *AdapterServiceTransport) ListProxies(ctx context.Context, req *pb.ListProxiesRequest) (*pb.ListProxiesResponse, error) {
	_, resp, err := t.list_proxies.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ListProxiesResponse), nil
}

func decodeAdapterServiceListProxiesRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.ListProxiesRequest)
	return param.ListProxiesRequest{Adapter: req.Adapter, Pager: common_param.Pager{Limit: req.Limit, Offset: req.Offset}, Params: req.Params}, nil
}

func encodeAdapterServiceListProxiesResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(param.ListProxiesResponse)
	proxies, _ := slices.Map[param.RawProxy, *pb.RawProxy](resp.Proxies, func(i int, p param.RawProxy) (*pb.RawProxy, error) {
		var support_protos []manager_pb.Proto
		for _, proto := range p.Proto {
			switch proto {
			case model.PROTO_HTTP:
				support_protos = append(support_protos, manager_pb.Proto_PROTO_HTTP)
			case model.PROTO_HTTPS:
				support_protos = append(support_protos, manager_pb.Proto_PROTO_HTTPS)
			case model.PROTO_WEBSOCKET:
				support_protos = append(support_protos, manager_pb.Proto_PROTO_WEBSOCKET)
			case model.PROTO_SOCKET:
				support_protos = append(support_protos, manager_pb.Proto_PROTO_SOCKET)
			}
		}
		use_config := manager_pb.UseConfig{
			User:     p.UseConfig.User,
			Password: p.UseConfig.Password,
			Psn:      p.UseConfig.Psn,
			Host:     p.UseConfig.Host,
			Port:     p.UseConfig.Port,
			Extra:    p.UseConfig.Extra,
		}
		proxy := &pb.RawProxy{Proto: support_protos, Ip: p.Ip, Port: p.Port, Ttl: p.Ttl, UseConfig: &use_config}
		if p.Attr != nil {
			proxy.Attr = &manager_pb.Attr{
				Tags: p.Attr.Tags,
			}
		}
		return proxy, nil
	})
	page := pb.ListProxiesResponse_Page{Limit: resp.Page.Limit, Offset: resp.Page.Offset, Total: resp.Page.Total}
	status := manager_pb.ResponseStatus{
		Code:    resp.Status.Code,
		Result:  resp.Status.Result,
		Message: resp.Status.Message,
	}
	var encode_resp pb.ListProxiesResponse = pb.ListProxiesResponse{
		Proxies: proxies,
		Page:    &page,
		Status:  &status,
	}
	return &encode_resp, nil
}
