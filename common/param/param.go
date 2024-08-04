package param

import "google.golang.org/grpc/codes"

type StatusResponse struct {
	Code    int32  `json:"code" example:"0"`
	Result  string `json:"result" example:"success"`
	Message string `json:"message" example:"success"`
}

func (r StatusResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"Status.Code", r.Code,
		"Status.Result", r.Result,
		"Status.Message", r.Message,
	)
}

type Pager struct {
	Limit  int64 `json:"limit,string" validate:"required" example:"10"`
	Offset int64 `json:"offset,string" validate:"required" example:"0"`
}

func (p Pager) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"Page.Limit", p.Limit,
		"Page.Offset", p.Offset,
	)
}

type PagerResponse struct {
	Pager
	Total int64 `json:"total,string" example:"100"`
	Count int64 `json:"count,string" example:"100"`
}

func (r PagerResponse) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"Page.Offset", r.Offset,
		"Page.Limit", r.Limit,
		"Page.Total", r.Total,
	)
}

var STATUS_OK = StatusResponse{
	Code:    int32(codes.OK),
	Result:  "success",
	Message: "success",
}
