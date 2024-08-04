package internal

import (
	"sync"

	server "github.com/WALL-EEEEEEE/proxy-service/gateway/internal/server"

	"github.com/go-gost/core/logger"
)

type Gateway struct {
	logger  logger.Logger
	servers []server.Server
}

func NewGateway(logger logger.Logger) *Gateway {
	return &Gateway{logger: logger}
}

func (g *Gateway) AddServer(s server.Server) {
	g.servers = append(g.servers, s)
}

func (g *Gateway) Serve() error {
	var wg sync.WaitGroup
	for _, s := range g.servers {
		serv := s
		wg.Add(1)
		go func() {
			defer wg.Done()
			serv.Serve()
		}()
	}
	wg.Wait()
	return nil
}
