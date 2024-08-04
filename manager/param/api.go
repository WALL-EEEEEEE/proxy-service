package param

import (
	"fmt"

	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
)

type AddApiRequest struct {
	Service        model.Service
	UpdateInterval float64
	Name           string
	ProviderId     string
}

func (r AddApiRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"AddApiRequest.Service", fmt.Sprintf("%+v", r.Service),
		"AddApiRequest.UpdateInterval", r.UpdateInterval,
		"AddApiRequest.Name", r.Name,
		"AddApiRequest.ProviderId", r.ProviderId,
	)
}

type AddApiResponse struct {
	common_param.StatusResponse
	Id string
}

func (r AddApiResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = r.StatusResponse.AppendKeyvals(keyvals)
	return append(keyvals,
		"AddApiResponse.Id", r.Id,
	)
}

type GetApiRequest struct {
	Id string
}

func (r GetApiRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"GetApiResponse.Id", r.Id,
	)
}

type GetApiResponse struct {
	common_param.StatusResponse
	Api model.ProxyApi
}

func (r GetApiResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = r.StatusResponse.AppendKeyvals(keyvals)
	return append(keyvals,
		"GetApiResponse.Api", fmt.Sprintf("%+v", r.Api),
	)
}

type GetApiByProviderRequest struct {
	ProviderId string
}

func (r GetApiByProviderRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"GetApiByProviderRequest.ProviderId", r.ProviderId,
	)
}

type GetApiByProviderResponse struct {
	common_param.StatusResponse
	Apis []model.ProxyApi
}

func (r GetApiByProviderResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	keyvals = r.StatusResponse.AppendKeyvals(keyvals)
	return append(keyvals,
		"GetApiByProviderRequest.Api.Length", len(r.Apis),
	)
}
