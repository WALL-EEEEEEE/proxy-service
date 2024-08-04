package manager

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/WALL-EEEEEEE/proxy-service/common"

	"github.com/WALL-EEEEEEE/proxy-service/manager/cache"
	conf "github.com/WALL-EEEEEEE/proxy-service/manager/config"
	ends "github.com/WALL-EEEEEEE/proxy-service/manager/endpoint"
	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	servs "github.com/WALL-EEEEEEE/proxy-service/manager/service"
	trans "github.com/WALL-EEEEEEE/proxy-service/manager/transport"

	log "github.com/sirupsen/logrus"

	protovalidate "github.com/bufbuild/protovalidate-go"
	gokit_grpc "github.com/go-kit/kit/transport/grpc"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var logger = log.StandardLogger()

var LoggingEndpointMiddleware = common.LoggingMiddleware

func startGrpcServer(ctx context.Context, conf *conf.Config, grpcPort string) {
	logger.Debugf("Config: %+v", conf)
	redis_cli, err := SetupRedis(conf)
	if err != nil {
		logger.Fatalf("failed to init redis! (error: %+v) ", err)
	}
	db, err := SetupDb(conf)
	if err != nil {
		logger.Fatalf("failed to init db! (error: %+v) ", err)
	}
	//proxy service
	proxy_store := cache.NewProxyStore(redis_cli)
	proxy_service := servs.NewProxyService(logger, proxy_store)
	proxy_service_end := ends.NewProxyServiceEndpoint(proxy_service)
	//add request auto logging
	proxy_service_end.ListProxies = LoggingEndpointMiddleware(logger, logger)(proxy_service_end.ListProxies)
	proxy_service_end.AddProxy = LoggingEndpointMiddleware(logger, logger)(proxy_service_end.AddProxy)
	proxy_service_end.UpdateProxy = LoggingEndpointMiddleware(logger, logger)(proxy_service_end.UpdateProxy)
	proxy_service_end.DeleteProxy = LoggingEndpointMiddleware(logger, logger)(proxy_service_end.DeleteProxy)
	proxy_service_end.GetProxy = LoggingEndpointMiddleware(logger, logger)(proxy_service_end.GetProxy)
	proxy_service_end.GetProxyByIp = LoggingEndpointMiddleware(logger, logger)(proxy_service_end.GetProxyByIp)

	proxy_service_grpc_server := trans.NewProxyServiceTransport(proxy_service_end, logger)

	// proxy provider service
	proxy_provider_service := servs.NewProxyProviderService(logger, db)
	proxy_provider_service_end := ends.NewProxyProviderServiceEndpoint(proxy_provider_service)
	//add request auto logging
	proxy_provider_service_end.GetProvider = LoggingEndpointMiddleware(logger, logger)(proxy_provider_service_end.GetProvider)
	proxy_provider_service_end.AddProvider = LoggingEndpointMiddleware(logger, logger)(proxy_provider_service_end.AddProvider)
	proxy_provider_service_end.UpdateProvider = LoggingEndpointMiddleware(logger, logger)(proxy_provider_service_end.UpdateProvider)
	proxy_provider_service_end.DeleteProvider = LoggingEndpointMiddleware(logger, logger)(proxy_provider_service_end.DeleteProvider)

	proxy_provider_service_grpc_server := trans.NewProxyProviderServiceTransport(proxy_provider_service_end, logger)

	//proxy api service
	proxy_api_service := servs.NewProxyApiService(db, redis_cli)
	proxy_api_service_end := ends.NewProxyApiServiceEndpoint(proxy_api_service)

	//add request auto logging
	proxy_api_service_end.GetApi = LoggingEndpointMiddleware(logger, logger)(proxy_api_service_end.GetApi)
	proxy_api_service_end.AddApi = LoggingEndpointMiddleware(logger, logger)(proxy_api_service_end.AddApi)
	proxy_api_service_end.DeleteApi = LoggingEndpointMiddleware(logger, logger)(proxy_api_service_end.UpdateApi)
	proxy_api_service_end.UpdateApi = LoggingEndpointMiddleware(logger, logger)(proxy_api_service_end.DeleteApi)

	proxy_api_service_grpc_server := trans.NewProxyApiServiceTransport(proxy_api_service_end, logger)

	// The gRPC listener mounts the Go kit gRPC server we created.
	grpcAddr := fmt.Sprintf(":%s", grpcPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatalf("failed to listen grpc port %s! (error: %+v) ", grpcAddr, err)
	}
	defer grpcListener.Close()
	logger.Infof("start grpc at %s ...", grpcAddr)

	//init the validator
	validator, err := protovalidate.New()
	if err != nil {
		logger.Fatalf("failed to init validator! (error: %+v) ", err)
	}
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(protovalidate_middleware.UnaryServerInterceptor(validator), gokit_grpc.Interceptor))

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthcheck)
	healthcheck.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	//register the proxy service grpc server
	pb.RegisterProxyServiceServer(grpcServer, proxy_service_grpc_server)
	//register the proxy provider service grpc server
	pb.RegisterProxyProviderServiceServer(grpcServer, proxy_provider_service_grpc_server)
	//register the proxy api service grpc server
	pb.RegisterProxyApiServiceServer(grpcServer, proxy_api_service_grpc_server)
	//enable reflection service on gRPC server.
	reflection.Register(grpcServer)

	grpcServer.Serve(grpcListener)
}

func startHttpServer(ctx context.Context, conf *conf.Config, grpcPort, httpPort string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := pb.RegisterProxyServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts); err != nil {
		logger.Fatalf("failed to register http handler for proxy service in the gateway: %v", err)
	}

	if err := pb.RegisterProxyProviderServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts); err != nil {
		logger.Fatalf("failed to register http handler for proxy provider service in the gateway: %v", err)
	}

	if err := pb.RegisterProxyApiServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts); err != nil {
		logger.Fatalf("failed to register http handler for proxy api service in the gateway: %v", err)
	}

	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: mux,
	}
	logger.Infof("start http/rest gateway at :%s ...", httpPort)
	srv.ListenAndServe()
}

func StartServer(conf *conf.Config, grpcPort, httpPort string) {
	ctx := context.Background()
	go startGrpcServer(ctx, conf, grpcPort)
	go startHttpServer(ctx, conf, grpcPort, httpPort)
	<-ctx.Done()
}
