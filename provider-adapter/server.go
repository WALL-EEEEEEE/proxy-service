package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/WALL-EEEEEEE/proxy-service/common"

	conf "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/config"
	ends "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/endpoint"
	pb "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/gen/adapter/v1"
	servs "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/service"
	trans "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/transport"

	protovalidate "github.com/bufbuild/protovalidate-go"
	gokit_grpc "github.com/go-kit/kit/transport/grpc"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var LoggingEndpointMiddleware = common.LoggingMiddleware

func startGrpcServer(ctx context.Context, logger *log.Logger, conf *conf.Config, grpcPort string) {
	logger.Infof("Config: %+v", conf)

	//bright data adapter service
	brightdata_service := servs.NewBrightDataAdapterService(logger, conf)
	//brightdata_service_mw := LoggingServiceMiddleware{logger: gokit_logger, next: brightdata_service}
	brightdata_service_end := ends.NewAdapterServiceEndpoint(brightdata_service)
	brightdata_service_end.ListProxies = RoutingMiddleware(brightdata_service)(brightdata_service_end.ListProxies)
	brightdata_service_end.ListProxies = LoggingEndpointMiddleware(logger, logger)(brightdata_service_end.ListProxies)
	brightdata_service_grpc_server := trans.NewAdapterServiceTransport(logger, brightdata_service_end)

	// The gRPC listener mounts the Go kit gRPC server we created.
	grpcAddr := fmt.Sprintf(":%s", grpcPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatalf("Failed to listen GRPC port %s! (Error: %+v) ", grpcAddr, err)
	}
	defer grpcListener.Close()
	logger.Infof("Start Grpc at %s ...", grpcAddr)

	//init the validator
	validator, err := protovalidate.New()
	if err != nil {
		logger.Fatalf("Failed to init validator! (Error: %+v) ", err)
	}
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(protovalidate_middleware.UnaryServerInterceptor(validator), gokit_grpc.Interceptor))

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthcheck)
	healthcheck.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	//enable reflection service on gRPC server.
	reflection.Register(grpcServer)

	//register the brightdata service grpc server
	pb.RegisterAdapterServiceServer(grpcServer, brightdata_service_grpc_server)

	grpcServer.Serve(grpcListener)
}

func startHttpServer(ctx context.Context, logger *log.Logger, conf *conf.Config, grpcPort, httpPort string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(10 * 1024 * 1024))}

	if err := pb.RegisterAdapterServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts); err != nil {
		logger.Fatalf("Failed to register HTTP handler for ProviderService in the  gateway: %v", err)
	}

	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: mux,
	}

	logger.Infof("Start HTTP/REST gateway at :%s ...", httpPort)
	srv.ListenAndServe()
}

func StartServer(logger *log.Logger, conf *conf.Config, grpcPort, httpPort string) {
	ctx := context.Background()
	go startGrpcServer(ctx, logger, conf, grpcPort)
	go startHttpServer(ctx, logger, conf, grpcPort, httpPort)
	<-ctx.Done()
}
