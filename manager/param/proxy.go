package param

import (
	"fmt"

	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"
	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
)

type ListProxiesRequest struct {
	common_param.Pager
	Filter   *pb.Filter
	ListMask []string
}

func (req ListProxiesRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = req.Pager.AppendKeyvals(keyvals)
	return append(keyvals,
		"ListProxiesRequest.Filter", req.Filter,
		"ListProxiesRequest.ListMask", req.ListMask,
	)
}

type ListProxiesResponse struct {
	common_param.StatusResponse
	common_param.PagerResponse
	ProxyList []model.Proxy
	ListMask  []string
}

func (resp ListProxiesResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = resp.StatusResponse.AppendKeyvals(keyvals)
	keyvals = resp.PagerResponse.AppendKeyvals(keyvals)
	return append(keyvals,
		"ListProxiesResponse.Proxies", len(resp.ProxyList),
		"ListProxiesResponse.ListMask", resp.ListMask,
	)
}

type AddProxyRequest struct {
	ProviderId string
	ApiId      string
	Proxy      model.Proxy
}

func (req AddProxyRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"AddProxyRequest.ApiId", req.ApiId,
		"AddProxyRequest.ProviderId", req.ProviderId,
		"AddProxyRequest.Proxy", fmt.Sprintf("%+v", req.Proxy),
	)
}

type AddProxyResponse struct {
	common_param.StatusResponse
	Id string
}

func (resp AddProxyResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = resp.StatusResponse.AppendKeyvals(keyvals)
	return append(keyvals,
		"AddProxyResponse.Id", resp.Id,
	)
}

type UpdateProxyRequest struct {
	Id         string
	Proxy      model.Proxy
	UpdateMask []string
}

func (req UpdateProxyRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"UpdateProxyRequest.Id", req.Id,
		"UpdateProxyRequest.Proxy", fmt.Sprintf("%+v", req.Proxy),
		"UpdateProxyRequest.UpdateMask", fmt.Sprintf("%+v", req.UpdateMask),
	)
}

type UpdateProxyResponse struct {
	common_param.StatusResponse
}

func (resp UpdateProxyResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	return resp.StatusResponse.AppendKeyvals(keyvals)
}

type GetProxyRequest struct {
	Id string
}

func (req GetProxyRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"GetProxyRequest.Id", req.Id,
	)
}

type GetProxyResponse struct {
	common_param.StatusResponse
	Proxy model.Proxy
}

func (resp GetProxyResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = resp.StatusResponse.AppendKeyvals(keyvals)
	return append(keyvals, "GetProxyResponse.Proxy", fmt.Sprintf("%+v", resp.Proxy))
}

type GetProxyByIpRequest struct {
	Ip string
}

func (req GetProxyByIpRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"GetProxyRequest.Ip", req.Ip,
	)
}

type GetProxyByIpResponse struct {
	common_param.StatusResponse
	Proxy model.Proxy
}

func (resp GetProxyByIpResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = resp.StatusResponse.AppendKeyvals(keyvals)
	return append(keyvals, "GetProxyByIpResponse.Proxy", fmt.Sprintf("%+v", resp.Proxy))
}
