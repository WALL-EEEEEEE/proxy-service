package internal

import (
	"context"
	"net"

	"github.com/WALL-EEEEEEE/proxy-service/gateway/log"

	"github.com/WALL-EEEEEEE/proxy-service/gateway/internal/handler"
	"github.com/WALL-EEEEEEE/proxy-service/gateway/internal/listener"

	"github.com/sirupsen/logrus"
)

type IServer interface {
	GetPort() int
	Serve() error
}

type Server struct {
	name     string
	logger   log.Logger
	handler  handler.Handler
	listener listener.Listener
}

type ServerOptions struct {
	name   string
	logger *log.Logger
}

type ServerOption func(opts *ServerOptions)

func NameServerOption(name string) ServerOption {
	return func(opts *ServerOptions) {
		opts.name = name
	}
}
func LogServerOption(logger *log.Logger) ServerOption {
	return func(opts *ServerOptions) {
		opts.logger = logger
	}
}

func NewServer(listener listener.Listener, handler handler.Handler, opts ...ServerOption) *Server {
	options := &ServerOptions{}
	for _, opt := range opts {
		opt(options)
	}
	var name string
	var logger log.Logger
	if options.name == "" {
		name = "server"
	} else {
		name = options.name
	}
	if options.logger == nil {
		logger = log.DefaultLogger
	} else {
		logger = *options.logger
	}

	return &Server{name: name, logger: logger, listener: listener, handler: handler}
}
func (s *Server) invokeHandle(conn net.Conn) {
	logger := s.logger.WithFields(logrus.Fields{
		"class":  s.name,
		"method": "invokeHandle",
	})
	defer conn.Close()
	ctx := context.Background()
	err := s.handler.Handle(ctx, conn)
	if err != nil {
		logger.Errorf("failed to handle connection: %v", err)
		return
	}
}

func (s *Server) Serve() {
	logger := s.logger.WithFields(logrus.Fields{
		"class":  s.name,
		"method": "Serve",
	})
	logger.Infof("listen on: %s", s.listener.Addr())
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			logger.Errorf("failed to accept connection from listener: %v", err)
			continue
		}
		go s.invokeHandle(conn)
	}
}

func (s *Server) GetPort() int {
	return s.listener.Port()
}
