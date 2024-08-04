package handler

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/gateway/log"
)

type HttpHandle func(context.Context, *HttpHandler, net.Conn, *http.Request) error

type HttpHandlerOptions struct {
	logger  *log.Logger
	handle  *HttpHandle
	timeout time.Duration
}

type HttpHandlerOption func(*HttpHandlerOptions)

func LoggerHttpHandlerOption(logger *log.Logger) HttpHandlerOption {
	return func(options *HttpHandlerOptions) {
		options.logger = logger
	}
}
func TimeoutHttpHandlerOption(timeout time.Duration) HttpHandlerOption {
	return func(options *HttpHandlerOptions) {
		options.timeout = timeout
	}
}
func HandleHttpHandlerOption(handle *HttpHandle) HttpHandlerOption {
	return func(options *HttpHandlerOptions) {
		options.handle = handle
	}
}
func defaultHandler(ctx context.Context, h *HttpHandler, conn net.Conn, r *http.Request) error {
	logger := h.Logger()
	logger.Warnf("no handle set")
	return nil
}

func NewHttpHandler(opts ...HttpHandlerOption) *HttpHandler {
	options := &HttpHandlerOptions{}
	for _, opt := range opts {
		opt(options)
	}
	h := &HttpHandler{options: *options}
	if options.handle == nil {
		h.handle = defaultHandler
	} else {
		h.handle = *options.handle
	}
	if options.logger == nil {
		h.logger = log.DefaultLogger
	} else {
		h.logger = *options.logger
	}
	return h
}

type HttpHandler struct {
	handle  HttpHandle
	logger  log.Logger
	options HttpHandlerOptions
}

func (h *HttpHandler) Logger() log.Logger {
	return h.logger
}

func (h *HttpHandler) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()
	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return err
	}
	defer req.Body.Close()
	return h.handle(ctx, h, conn, req)
}
