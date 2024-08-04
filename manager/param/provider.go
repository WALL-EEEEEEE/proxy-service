package param

import (
	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
)

type GetProviderRequest struct {
	Id string
}

func (r GetProviderRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"GetProviderRequest.Id", r.Id,
	)
}

type GetProviderResponse struct {
	common_param.StatusResponse
	Provider model.ProxyProvider
}

func (r GetProviderResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = r.StatusResponse.AppendKeyvals(keyvals)
	return append(keyvals,
		"GetProviderResponse.Provider", r.Provider,
	)
}

type AddProviderRequest struct {
	Name string
}

func (r AddProviderRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"AddProviderResponse.Name", r.Name,
	)
}

type AddProviderResponse struct {
	common_param.StatusResponse
	Id string
}

func (r AddProviderResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = r.StatusResponse.AppendKeyvals(keyvals)
	return append(keyvals,
		"AddProviderResponse.Id", r.Id,
	)
}

type ListProviderRequest struct {
	common_param.Pager
}

func (r ListProviderRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return r.Pager.AppendKeyvals(keyvals)
}

type ListProviderResponse struct {
	common_param.StatusResponse
	ProviderList []model.ProxyProvider
}

func (r ListProviderResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = r.StatusResponse.AppendKeyvals(keyvals)
	return append(keyvals,
		"ListProviderResponse.Provider.Length", len(r.ProviderList),
	)
}
