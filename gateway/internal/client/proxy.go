package client

import (
	"context"
	"fmt"

	log "github.com/WALL-EEEEEEE/proxy-service/gateway/log"
	managerv1_pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var (
	default_masks = []string{"id", "proto", "ip", "port", "status", "provider", "api", "attr", "created_at", "updated_at", "checked_at", "expire_time", "use_config"}
	base_filters  = []*managerv1_pb.Filter{
		ConstructPropertyFilter("status", managerv1_pb.PropertyFilter_EQUAL, managerv1_pb.Status_STATUS_CHECKED.String()),
	}
)

type ProxyClientOptions struct {
	logger *log.Logger
	ctx    *context.Context
}

type ProxyClientOption func(*ProxyClientOptions)

func LogProxyClientOption(logger *log.Logger) ProxyClientOption {
	return func(options *ProxyClientOptions) {
		options.logger = logger
	}
}
func CtxProxyClientOption(ctx *context.Context) ProxyClientOption {
	return func(options *ProxyClientOptions) {
		options.ctx = ctx
	}
}

type ListProxiesOptions struct {
	filters []*managerv1_pb.Filter
	masks   []string
}
type ListProxiesOption func(*ListProxiesOptions)

func FilterListProxiesOption(filters ...*managerv1_pb.Filter) ListProxiesOption {
	return func(options *ListProxiesOptions) {
		options.filters = append(options.filters, filters...)
	}
}

func MaskListProxiesOption(masks ...string) ListProxiesOption {
	return func(options *ListProxiesOptions) {
		options.masks = append(options.masks, masks...)
	}
}

type ProxyClient struct {
	grpc_addr   string
	logger      log.Logger
	grpc_client managerv1_pb.ProxyServiceClient
}

func NewProxyClient(grpc_addr string, opts ...ProxyClientOption) (*ProxyClient, error) {
	options := &ProxyClientOptions{}
	for _, opt := range opts {
		opt(options)
	}
	var (
		ctx    context.Context
		logger log.Logger
	)
	if options.ctx == nil {
		ctx = context.Background()
	} else {
		ctx = *options.ctx
	}
	if options.logger == nil {
		logger = log.DefaultLogger
	} else {
		logger = *options.logger
	}
	client := &ProxyClient{logger: logger, grpc_addr: grpc_addr}
	if err := client.initGrpcClient(ctx, grpc_addr); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *ProxyClient) initGrpcClient(ctx context.Context, addr string) error {
	serviceConfig := grpc.WithDefaultServiceConfig(`
	{
		"loadBalancingPolicy": "round_robin",
		"healthCheckConfig": {
			"serviceName": "proxy.manager.v1.ProxyService"
		}
	}`)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), serviceConfig)
	if err != nil {
		return fmt.Errorf("failed to dial grpc server: %v", err)
	}
	grpc_client := managerv1_pb.NewProxyServiceClient(conn)
	c.grpc_client = grpc_client
	return nil
}

func (s *ProxyClient) ListProxies(ctx context.Context, limit int, offset int, opts ...ListProxiesOption) ([]*managerv1_pb.Proxy, error) {
	var proxies []*managerv1_pb.Proxy = make([]*managerv1_pb.Proxy, 0)
	// Get proxies
	req, err := ConstructListProxiesRequest(limit, offset, opts...)
	if err != nil {
		return nil, err
	}
	resp, err := s.grpc_client.ListProxies(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Status.Code != 0 {
		return nil, fmt.Errorf("failed to list proxies: %v", resp.Status)
	}
	if resp.Count < 1 {
		return nil, nil
	}
	proxies = append(proxies, resp.GetProxyList()...)
	return proxies, nil
}

func (s *ProxyClient) GetAddr() string {
	return s.grpc_addr
}

func ConstructListProxiesRequest(limit int, offset int, opts ...ListProxiesOption) (*managerv1_pb.ListProxiesRequest, error) {
	options := &ListProxiesOptions{}
	for _, opt := range opts {
		opt(options)
	}
	req := &managerv1_pb.ListProxiesRequest{
		Query: &managerv1_pb.Query{
			Limit:  int64(limit),
			Offset: int32(offset),
		},
	}
	var filters []*managerv1_pb.Filter
	filters = append(filters, base_filters...)
	if options.filters != nil {
		filters = append(filters, options.filters...)

	}
	req.Query.Filter = &managerv1_pb.Filter{
		FilterType: &managerv1_pb.Filter_CompositeFilter{
			CompositeFilter: &managerv1_pb.CompositeFilter{
				Op:      managerv1_pb.CompositeFilter_AND,
				Filters: filters,
			},
		},
	}
	var (
		masks []string
	)
	if options.masks != nil {
		masks = options.masks
	} else {
		masks = default_masks
	}
	filed_masks, err := fieldmaskpb.New(&managerv1_pb.Proxy{}, masks...)
	if err != nil {
		return nil, err
	}
	req.Fields = filed_masks
	return req, nil
}

func ConstructPropertyFilter(name string, op managerv1_pb.PropertyFilter_Operator, value string) *managerv1_pb.Filter {
	return &managerv1_pb.Filter{
		FilterType: &managerv1_pb.Filter_PropertyFilter{
			PropertyFilter: &managerv1_pb.PropertyFilter{
				Property: &managerv1_pb.PropertyReference{
					Name: name,
				},
				Op:    op,
				Value: value,
			},
		},
	}
}
