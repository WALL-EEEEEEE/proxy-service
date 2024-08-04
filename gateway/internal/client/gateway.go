package client

import (
	"context"
	"fmt"

	log "github.com/WALL-EEEEEEE/proxy-service/gateway/log"
	managerv1_pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GatewayClientOptions struct {
	logger *log.Logger
	ctx    *context.Context
}

type GatewayClientOption func(*GatewayClientOptions)

func LogGatewayClientOption(logger *log.Logger) GatewayClientOption {
	return func(options *GatewayClientOptions) {
		options.logger = logger
	}
}
func CtxGatewayClientOption(ctx *context.Context) GatewayClientOption {
	return func(options *GatewayClientOptions) {
		options.ctx = ctx
	}
}

func NewGatewayClient(grpc_addr string, opts ...GatewayClientOption) (*GatewayClient, error) {
	options := &GatewayClientOptions{}
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
	client := &GatewayClient{logger: logger}
	if err := client.initGrpcClient(ctx, grpc_addr); err != nil {
		return nil, err
	}
	return client, nil
}

// TODO: implement GatewayClient interacting with gateway service
type GatewayClient struct {
	logger      log.Logger
	grpc_client managerv1_pb.GatewayServiceClient
}

func (c *GatewayClient) initGrpcClient(ctx context.Context, addr string) error {
	serviceConfig := grpc.WithDefaultServiceConfig(`
	{
		"loadBalancingPolicy": "round_robin",
		"healthCheckConfig": {
			"serviceName": "proxy.manager.v1.GatewayService"
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
