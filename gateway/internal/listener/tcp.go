package listener

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/WALL-EEEEEEE/proxy-service/gateway/log"

	"github.com/sirupsen/logrus"
)

type TcpListenerOptions struct {
	logger *log.Logger
	ctx    context.Context
}

func LoggerTcpListenerOption(logger *log.Logger) TcpListenerOption {
	return func(options *TcpListenerOptions) {
		options.logger = logger
	}
}

type TcpListenerOption func(*TcpListenerOptions)

type TcpListener struct {
	ln     net.Listener
	logger log.Logger
	opts   *TcpListenerOptions
}

func NewTcpListener(port int, opts ...TcpListenerOption) (*TcpListener, error) {
	options := &TcpListenerOptions{}
	for _, opt := range opts {
		opt(options)
	}
	var ctx context.Context
	if options.ctx != nil {
		ctx = options.ctx
	} else {
		ctx = context.Background()
	}
	lc := net.ListenConfig{}
	addr := fmt.Sprintf(":%d", port)
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	var logger log.Logger
	if options.logger != nil {
		logger = *options.logger
	} else {
		logger = log.DefaultLogger
	}
	tcp_ln := &TcpListener{ln: ln, opts: options, logger: logger}
	return tcp_ln, nil
}

func (l *TcpListener) Accept() (net.Conn, error) {
	_ = l.logger.WithFields(logrus.Fields{
		"class":  "TcpListener",
		"Method": "Accept",
	})
	return l.ln.Accept()
}

func (l *TcpListener) Addr() string {
	return l.ln.Addr().String()
}

func (l *TcpListener) Port() int {
	addr := l.ln.Addr().String()
	parts := strings.Split(addr, ":")
	var port int = -1
	var err error
	if len(parts) > 1 {
		port, err = strconv.Atoi(parts[0])
	}
	if err != nil {
		panic(err)
	}
	return port
}
