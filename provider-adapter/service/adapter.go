package service

import (
	"context"

	"github.com/WALL-EEEEEEE/proxy-service/provider-adapter/param"
)

type IProxyAdapterService interface {
	ListProxies(context.Context, int64, int64, map[string]string) (int64, int64, []param.RawProxy, error)
	GetName() string
}
