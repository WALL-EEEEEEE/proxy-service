package util

import (
	"fmt"
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/common"
	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"

	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/WALL-EEEEEEE/proxy-service/manager/param"

	"github.com/bobg/go-generics/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	FILTERSET_AND    = common.FILTERSET_AND
	FILTERSET_OR     = common.FILTERSET_OR
	FILTER_EQUAL     = common.FILTER_EQUAL
	FILTER_NOT_EQUAL = common.FILTER_NOT_EQUAL
)

type FilterSetOperator = common.FilterSetOperator
type Filter = common.Filter[any]
type Pager = common_param.Pager

func FilterFromPb(filter *pb.Filter) (filters []Filter, err error) {
	if filter == nil {
		return nil, nil
	}
	switch filter.GetFilterType().(type) {
	case *pb.Filter_CompositeFilter:
		for _, f := range filter.GetCompositeFilter().GetFilters() {
			sub_filters, err := FilterFromPb(f)
			filters = append(filters, sub_filters...)
			if err != nil {
				return nil, err
			}
		}
	case *pb.Filter_PropertyFilter:
		_filter := filter.GetPropertyFilter()
		_filter_op := _filter.GetOp()
		switch _filter_op {
		case pb.PropertyFilter_EQUAL:
			query_filter := Filter{
				Op:    &FILTER_EQUAL,
				Value: _filter.GetValue(),
				Name:  _filter.GetProperty().GetName(),
				SetOp: &FILTERSET_AND,
			}
			filters = append(filters, query_filter)
		case pb.PropertyFilter_NOT_EQUAL:
			query_filter := Filter{
				Op:    &FILTER_NOT_EQUAL,
				Value: _filter.GetValue(),
				Name:  _filter.GetProperty().GetName(),
				SetOp: &FILTERSET_AND,
			}
			filters = append(filters, query_filter)
		default:
			err = fmt.Errorf("invalid query filter operator: %s", _filter.GetOp().String())
			return nil, err
		}
	}
	return filters, nil
}

func ProtosFromPb(proto []pb.Proto) []model.PROTO {
	ret_proto, _ := slices.Map(proto, func(_ int, proto pb.Proto) (model.PROTO, error) {
		switch proto {
		case pb.Proto_PROTO_HTTP:
			return model.PROTO_HTTP, nil
		case pb.Proto_PROTO_HTTPS:
			return model.PROTO_HTTPS, nil
		case pb.Proto_PROTO_WEBSOCKET:
			return model.PROTO_WEBSOCKET, nil
		default:
			return model.PROTO{Value: "UNKNOW"}, nil
		}
	})
	return ret_proto
}

func PbFromStatus(status model.STATUS) pb.Status {
	switch status {
	case model.STATUS_CREATED:
		return pb.Status_STATUS_CREATED
	case model.STATUS_CHECKED:
		return pb.Status_STATUS_CHECKED
	default:
		return pb.Status_STATUS_UNSPECIFIED
	}
}

func PbFromProtos(protos []model.PROTO) []pb.Proto {
	ret_proto, _ := slices.Map[model.PROTO, pb.Proto](protos, func(_ int, proto model.PROTO) (pb.Proto, error) {
		switch proto {
		case model.PROTO_HTTP:
			return pb.Proto_PROTO_HTTP, nil
		case model.PROTO_HTTPS:
			return pb.Proto_PROTO_HTTPS, nil
		case model.PROTO_WEBSOCKET:
			return pb.Proto_PROTO_WEBSOCKET, nil
		default:
			return pb.Proto_PROTO_UNSPECIFIED, nil
		}
	})
	return ret_proto
}

func PbFromUseConfig(use_config *model.UseConfig) *pb.UseConfig {
	if use_config == nil {
		return nil
	}
	return &pb.UseConfig{
		Psn:      use_config.Psn,
		Host:     use_config.Host,
		Port:     use_config.Port,
		User:     use_config.User,
		Password: use_config.Password,
		Extra:    use_config.Extra,
	}
}

func PbFromAttr(attr *model.Attr) *pb.Attr {
	if attr == nil {
		return nil
	}
	return &pb.Attr{
		Country:      attr.Country,
		City:         attr.City,
		Organization: attr.Organization,
		Location:     attr.Location,
		Region:       attr.Region,
		Latency:      attr.Latency,
		Stability:    attr.Stability,
		Availiable:   attr.Availiable,
		Anonymous:    attr.Anonymous,
		Tags:         attr.Tags,
	}
}

func PbFromProxy(proxy *model.Proxy) *pb.Proxy {
	if proxy == nil {
		return nil
	}
	ret_proxy := pb.Proxy{
		Id:         proxy.Id,
		Ip:         proxy.Ip,
		Port:       proxy.Port,
		ProviderId: proxy.ProviderId,
		ApiId:      proxy.ApiId,
		Provider:   proxy.Provider,
		Api:        proxy.Api,
		Proto:      PbFromProtos(proxy.Proto),
		Attr:       PbFromAttr(proxy.Attr),
		Status:     PbFromStatus(proxy.Status),
		UseConfig:  PbFromUseConfig(proxy.UseConfig),
	}

	if proxy.UpdatedAt != nil {
		ret_proxy.UpdatedAt = timestamppb.New(*proxy.UpdatedAt)
	}
	if proxy.CreatedAt != nil {
		ret_proxy.CreatedAt = timestamppb.New(*proxy.CreatedAt)
	}

	//Ttl precedes Expiration if Ttl equate to -1
	if proxy.Ttl == -1 {
		ret_proxy.Expiration = &pb.Proxy_Ttl{Ttl: -1}
	} else {
		//Expiration precedes Ttl while ExpiredAt is set and Ttl isn't equal to -1, otherwise caculate Expiration based CreatedAt and Ttl
		if !(proxy.ExpiredAt == nil || proxy.ExpiredAt.IsZero()) {
			ret_proxy.Expiration = &pb.Proxy_ExpireTime{ExpireTime: timestamppb.New(*proxy.ExpiredAt)}
		} else {
			expired_at := timestamppb.New((*proxy.CreatedAt).Add(time.Duration(proxy.Ttl) * time.Second))
			ret_proxy.Expiration = &pb.Proxy_ExpireTime{ExpireTime: expired_at}
		}
	}

	return &ret_proxy
}

func UseConfigFromPb(use_config *pb.UseConfig) *model.UseConfig {
	if use_config == nil {
		return nil
	}
	return &model.UseConfig{
		Psn:      use_config.Psn,
		Host:     use_config.Host,
		Port:     use_config.Port,
		User:     use_config.User,
		Password: use_config.Password,
		Extra:    use_config.Extra,
	}
}

func AttrFromPb(attr *pb.Attr) *model.Attr {
	if attr == nil {
		return nil
	}
	return &model.Attr{
		Country:      attr.Country,
		City:         attr.City,
		Organization: attr.Organization,
		Location:     attr.Location,
		Region:       attr.Region,
		Latency:      attr.Latency,
		Stability:    attr.Stability,
		Availiable:   attr.Availiable,
		Anonymous:    attr.Anonymous,
		Tags:         attr.Tags,
	}
}

func StatusFromPb(status pb.Status) model.STATUS {
	switch status {
	case pb.Status_STATUS_CREATED:
		return model.STATUS_CREATED
	case pb.Status_STATUS_CHECKED:
		return model.STATUS_CHECKED
	default:
		return model.STATUS_UNSPECIFIED
	}
}

func ExpirationFromPb(proxy *pb.Proxy, ret_proxy *model.Proxy) {
	switch proxy.Expiration.(type) {
	case *pb.Proxy_Ttl:
		ttl := proxy.GetTtl()
		if ttl != -1 {
			expired_at := proxy.CreatedAt.AsTime().Add(time.Duration(ttl) * time.Second)
			ret_proxy.ExpiredAt = &expired_at
		}
		ret_proxy.Ttl = ttl
	case *pb.Proxy_ExpireTime:
		expired_at := proxy.GetExpireTime().AsTime()
		ret_proxy.ExpiredAt = &expired_at
		ret_proxy.Ttl = int64(proxy.GetExpireTime().AsTime().Sub(proxy.CreatedAt.AsTime()).Seconds())
	}
}

func ProxyFromPb(proxy *pb.Proxy) *model.Proxy {
	if proxy == nil {
		return nil
	}
	ret_proxy := model.Proxy{
		Ip:         proxy.Ip,
		Port:       proxy.Port,
		ProviderId: proxy.ProviderId,
		ApiId:      proxy.ApiId,
		Provider:   proxy.Provider,
		Api:        proxy.Api,
		Proto:      ProtosFromPb(proxy.Proto),
		Attr:       AttrFromPb(proxy.Attr),
		Status:     StatusFromPb(proxy.Status),
		UseConfig:  UseConfigFromPb(proxy.UseConfig),
	}
	if proxy.UpdatedAt == nil {
		ret_proxy.UpdatedAt = nil
	} else {
		_updated_at := proxy.UpdatedAt.AsTime()
		ret_proxy.UpdatedAt = &_updated_at
	}
	if proxy.CreatedAt == nil {
		ret_proxy.CreatedAt = nil
	} else {
		_created_at := proxy.CreatedAt.AsTime()
		ret_proxy.CreatedAt = &_created_at
	}
	ExpirationFromPb(proxy, &ret_proxy)
	return &ret_proxy
}

func UpdateProxyRequestFromPb(req *pb.UpdateProxyRequest) *param.UpdateProxyRequest {
	var ret_req param.UpdateProxyRequest = param.UpdateProxyRequest{
		Id:         req.Id,
		Proxy:      *ProxyFromPb(req.Proxy),
		UpdateMask: req.Fields.Paths,
	}
	return &ret_req
}

func AddProxyRequestFromPb(req *pb.AddProxyRequest) *param.AddProxyRequest {
	var ret_req param.AddProxyRequest = param.AddProxyRequest{
		ProviderId: req.ProviderId,
		ApiId:      req.ApiId,
		Proxy:      *ProxyFromPb(req.Proxy),
	}
	return &ret_req
}
