package service

// Proxy provides operations on proxy.
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	conf "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/config"
	"github.com/WALL-EEEEEEE/proxy-service/provider-adapter/param"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	retry_http "github.com/hashicorp/go-retryablehttp"
)

const bright_data_ips_api = "https://api.brightdata.com/zone/route_ips?zone=%s"
const bright_data_zone_status_api = "https://api.brightdata.com/zone/status?zone=%s"
const bright_data_zone_info_api = "https://api.brightdata.com/zone/info?zone=%s"

type BrightDataAdapterService struct {
	IProxyAdapterService
	logger *log.Logger
	conf   *conf.Config
	client *http.Client
}

func NewBrightDataAdapterService(logger *log.Logger, conf *conf.Config) *BrightDataAdapterService {
	retryClient := retry_http.NewClient()
	retryClient.RetryMax = 3
	client := retryClient.StandardClient()
	client.Timeout = time.Duration(10) * time.Second
	return &BrightDataAdapterService{
		logger: logger,
		conf:   conf,
		client: retryClient.StandardClient(),
	}
}

func (s BrightDataAdapterService) non_gateway_proxies(ctx context.Context, offset int64, limit int64, params map[string]string) ([]param.RawProxy, error) {
	zone, ok := params["zone"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("params[zone] is required for adapter %s", s.GetName()))
	}
	exist, err := s.exists_zone(ctx, zone)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("failed to check existence of zone [%s] for adapter %s (err: %+v)", zone, s.GetName(), err.Error()))
	}
	if !exist {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("zone [%s] doesn't exist for adapter %s", zone, s.GetName()))
	}
	proxies, err := s.get_zone_ips(ctx, zone, params)
	if err != nil {
		return nil, err
	}
	return proxies, nil
}

func (s BrightDataAdapterService) gateway_proxies(ctx context.Context, offset int64, limit int64, params map[string]string) ([]param.RawProxy, error) {
	zone, ok := params["zone"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("params[zone] is required for adapter %s", s.GetName()))
	}
	exist, err := s.exists_zone(ctx, zone)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("failed to check existence of zone [%s] for adapter %s (err: %+v)", zone, s.GetName(), err.Error()))
	}
	if !exist {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("zone [%s] doesn't exist for adapter %s", zone, s.GetName()))
	}
	user, ok := params["user"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("params[user] is required for adapter %s", s.GetName()))
	}
	password, ok := params["password"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("params[password] is required for adapter %s", s.GetName()))
	}
	var proxy_tags []string = []string{"gateway"}
	tags, ok := params["tags"]
	if ok && len(tags) > 0 {
		proxy_tags = append(proxy_tags, strings.Split(tags, ",")...)
	}
	var proxies []param.RawProxy
	gateway := fmt.Sprintf("%s:%d", s.conf.BrightData.Gateway.Host, s.conf.BrightData.Gateway.Port)
	psn := fmt.Sprintf("http://%s:%s@%s", user, password, gateway)
	ip := fmt.Sprintf("brightdata_%s", zone)
	proxy := param.RawProxy{
		Ip:    ip,
		Proto: []model.PROTO{model.PROTO_HTTP, model.PROTO_HTTPS, model.PROTO_WEBSOCKET},
		Port:  -1,
		Ttl:   -1,
		UseConfig: &model.UseConfig{
			Psn:      psn,
			Host:     s.conf.BrightData.Gateway.Host,
			Port:     s.conf.BrightData.Gateway.Port,
			User:     user,
			Password: password,
			Extra: map[string]string{
				"zone": zone,
			},
		},
		Attr: &model.Attr{
			Tags: proxy_tags,
		},
	}

	proxies = append(proxies, proxy)
	return proxies, nil
}

func (s BrightDataAdapterService) ListProxies(ctx context.Context, offset int64, limit int64, params map[string]string) (ret_total int64, ret_offset int64, ret_proxies []param.RawProxy, err error) {
	ret_total = 0
	ret_offset = offset
	ret_proxies = nil
	if params == nil {
		return 0, 0, nil, status.Error(codes.InvalidArgument, fmt.Sprintf("params is required for adapter %s", s.GetName()))
	}
	kind, ok := params["kind"]
	if !ok {
		return 0, 0, nil, status.Error(codes.InvalidArgument, fmt.Sprintf(`params["kind"] is required for adapter %s`, s.GetName()))
	}
	switch kind {
	case "GATEWAY":
		ret_proxies, err = s.gateway_proxies(ctx, offset, limit, params)
	case "NON_GATEWAY":
		ret_proxies, err = s.non_gateway_proxies(ctx, offset, limit, params)
	default:
		return 0, 0, nil, status.Error(codes.InvalidArgument, fmt.Sprintf(`invalid params["kind"] value "%s"`, kind))
	}
	if err != nil {
		return 0, 0, nil, status.Error(codes.Unknown, err.Error())
	}
	ret_total = int64(len(ret_proxies))
	ret_offset = ret_total
	return
}

func (s BrightDataAdapterService) GetName() string {
	return "bright_data"
}

func (s BrightDataAdapterService) exists_zone(ctx context.Context, zone string) (bool, error) {
	url := fmt.Sprintf(bright_data_zone_status_api, zone)
	http_req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = status.Error(codes.Unknown, fmt.Sprintf("bright data zone status api error, %s", err.Error()))
		return false, err
	}
	auth_header := "Bearer " + s.conf.BrightData.Token
	http_req.Header.Add("Authorization", auth_header)
	resp, err := s.client.Do(http_req)
	if err != nil {
		err = status.Error(codes.Unknown, fmt.Sprintf("bright data zone status api error, %s", err.Error()))
		return false, err
	}
	defer resp.Body.Close()
	resp_content, err := io.ReadAll(resp.Body)
	if err != nil {
		err = status.Error(codes.Unknown, fmt.Sprintf("bright data zone status api error, %s", err.Error()))
		return false, err
	}
	if len(resp_content) < 1 {
		err = status.Error(codes.Unknown, "bright data zone status api response returned is  emtpy")
		return false, err
	}
	status_resp := struct {
		Status string `json:"status,omitempty"`
	}{}
	err = json.Unmarshal(resp_content, &status_resp)
	if err != nil {
		err = status.Error(codes.Unknown, "bright data zone status api response invalid")
		return false, err
	}
	if status_resp.Status != "active" {
		return false, nil
	}
	return true, nil
}

func (s BrightDataAdapterService) get_zone_ips(ctx context.Context, zone string, params map[string]string) ([]param.RawProxy, error) {
	url := fmt.Sprintf(bright_data_ips_api, zone)
	user, ok := params["user"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("params[user] is required for adapter %s", s.GetName()))
	}
	password, ok := params["password"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("params[password] is required for adapter %s", s.GetName()))
	}
	city, ok := params["city"]
	if ok {
		url += fmt.Sprintf("&city=%s", city)
	}
	http_req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = status.Error(codes.Unknown, fmt.Sprintf("bright data api error, %s", err.Error()))
		return nil, err
	}
	auth_header := "Bearer " + s.conf.BrightData.Token
	http_req.Header.Add("Authorization", auth_header)
	resp, err := s.client.Do(http_req)
	if err != nil {
		err = status.Error(codes.Unknown, fmt.Sprintf("bright data api error, %s", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()
	resp_content, err := io.ReadAll(resp.Body)
	if err != nil {
		err = status.Error(codes.Unknown, fmt.Sprintf("bright data api error, %s", err.Error()))
		return nil, err
	}
	if len(resp_content) < 1 {
		err = status.Error(codes.Unknown, "bright data api response returned is  emtpy")
		return nil, err
	}
	var proxy_tags []string = []string{"ip"}
	tags, ok := params["tags"]
	if ok && len(tags) > 0 {
		proxy_tags = append(proxy_tags, strings.Split(tags, ",")...)
	}
	resp_proxies := strings.Split(string(resp_content), "\n")
	var proxies []param.RawProxy
	for _, proxy := range resp_proxies {
		if len(proxy) < 1 {
			continue
		}
		gateway := fmt.Sprintf("%s:%d", s.conf.BrightData.Gateway.Host, s.conf.BrightData.Gateway.Port)
		psn := fmt.Sprintf("http://%s-ip-%s:%s@%s", user, proxy, password, gateway)
		user := fmt.Sprintf("%s-ip-%s", user, proxy)
		proxy := param.RawProxy{
			Ip:    proxy,
			Proto: []model.PROTO{model.PROTO_HTTP, model.PROTO_HTTPS, model.PROTO_WEBSOCKET},
			Port:  -1,
			Ttl:   -1,
			UseConfig: &model.UseConfig{
				Psn:      psn,
				Host:     s.conf.BrightData.Gateway.Host,
				Port:     s.conf.BrightData.Gateway.Port,
				User:     user,
				Password: password,
				Extra: map[string]string{
					"zone": zone,
				},
			},
		}
		if len(proxy_tags) > 0 {
			proxy.Attr = &model.Attr{
				Tags: proxy_tags,
			}
		}
		proxies = append(proxies, proxy)
		s.logger.Debugf("bright data api response proxy: %#v", proxy)
	}
	return proxies, nil
}
