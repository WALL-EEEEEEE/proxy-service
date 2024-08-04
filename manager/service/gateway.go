package service

// Proxy provides operations on proxy.
import (
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type IGatewayService interface {
}

type GatewayService struct {
	logger *logrus.Logger
}

func NewGatewayService(logger *logrus.Logger, redis_cli *redis.Client) *GatewayService {
	return nil
}
