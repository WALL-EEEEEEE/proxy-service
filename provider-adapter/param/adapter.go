package param

import (
	common "github.com/WALL-EEEEEEE/proxy-service/common/param"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
)

type RawProxy struct {
	Proto     []model.PROTO
	Ip        string
	Ttl       int64            `json:",string"`
	Port      int64            `json:",string"`
	Attr      *model.Attr      `json:"attr"`
	UseConfig *model.UseConfig `json:"use_config,omitempty"`
}

type ListProxiesRequest struct {
	Adapter string            `json:"adapter,omitempty"`
	Params  map[string]string `json:"params" validate:"required" example:"{}"`
	common.Pager
}

func (r ListProxiesRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = r.Pager.AppendKeyvals(keyvals)
	return append(keyvals,
		"ListProxiesRequest.Adapter", r.Adapter,
		"ListProxiesRequest.Params", r.Params,
		"ListProxiesRequest.Pager", r.Pager,
	)
}

type ListProxiesResponse struct {
	Status  common.StatusResponse `json:"status"`
	Page    common.PagerResponse  `json:"page"`
	Proxies []RawProxy
}

func (r ListProxiesResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = r.Status.AppendKeyvals(keyvals)
	keyvals = r.Page.AppendKeyvals(keyvals)
	return append(keyvals,
		"ListProxiesResponse.Proxy.Length", len(r.Proxies),
	)
}
