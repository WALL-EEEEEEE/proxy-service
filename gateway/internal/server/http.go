package internal

import (
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/gateway/internal/handler"
	listener "github.com/WALL-EEEEEEE/proxy-service/gateway/internal/listener"
	log "github.com/WALL-EEEEEEE/proxy-service/gateway/log"
)

type HttpProxyServerOptions struct {
	logger        *log.Logger
	handle        *handler.HttpHandle
	DailTimetout  time.Duration
	HandleTimeout time.Duration
}

type HttpProxyServerOption func(*HttpProxyServerOptions)

func LogHttpProxyServerOption(logger *log.Logger) HttpProxyServerOption {
	return func(options *HttpProxyServerOptions) {
		options.logger = logger
	}
}
func HandleTimeoutHttpProxyServerOption(timeout time.Duration) HttpProxyServerOption {
	return func(options *HttpProxyServerOptions) {
		options.HandleTimeout = timeout
	}
}
func HandleHttpProxyServerOption(handle handler.HttpHandle) HttpProxyServerOption {
	return func(options *HttpProxyServerOptions) {
		options.handle = &handle
	}
}

func DailTimeoutHttpProxyServerOption(timeout time.Duration) HttpProxyServerOption {
	return func(options *HttpProxyServerOptions) {
		options.DailTimetout = timeout
	}
}

func NewHttpProxyServer(port int, opts ...HttpProxyServerOption) (serv *Server, err error) {
	options := &HttpProxyServerOptions{}
	for _, opt := range opts {
		opt(options)
	}
	ln, err := listener.NewTcpListener(port,
		listener.LoggerTcpListenerOption(options.logger),
	)
	if err != nil {
		return nil, err
	}
	hd := handler.NewHttpHandler(
		handler.LoggerHttpHandlerOption(options.logger),
		handler.TimeoutHttpHandlerOption(options.HandleTimeout),
		handler.HandleHttpHandlerOption(options.handle),
	)
	serv = NewServer(ln, hd,
		LogServerOption(options.logger),
	)
	return serv, nil
}
