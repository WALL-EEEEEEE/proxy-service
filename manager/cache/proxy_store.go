package cache

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/common"

	"github.com/WALL-EEEEEEE/proxy-service/manager/event"
	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	redisearch "github.com/WALL-EEEEEEE/proxy-service/manager/util/redisearch"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	log "github.com/sirupsen/logrus"
)

type Proxy model.Proxy

func (p Proxy) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Proxy) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	return nil
}

type StoreOption struct {
	Schemas []redisearch.Schema //index for Cache Store
}
type SetOp = string

const (
	SETOP_CRETATE SetOp = "CREATE"
	SETOP_UPDATE  SetOp = "UPDATE"
)

type SetOption struct {
	Operation SetOp
}

var DefaultSetOption = SetOption{
	SETOP_CRETATE,
}

var (
	DefaultStoreOption StoreOption = StoreOption{
		Schemas: []redisearch.Schema{
			redisearch.NewSchema("proto", redisearch.SCHEMA_KIND_TAG, redisearch.AliasSchemaOption("proto")),
			redisearch.NewSchema("status", redisearch.SCHEMA_KIND_TAG, redisearch.AliasSchemaOption("status")),
			redisearch.NewSchema("attr.tags", redisearch.SCHEMA_KIND_TAG, redisearch.AliasSchemaOption("tags")),
			redisearch.NewSchema("id", redisearch.SCHEMA_KIND_TAG, redisearch.AliasSchemaOption("id")),
			redisearch.NewSchema("ip", redisearch.SCHEMA_KIND_TAG, redisearch.AliasSchemaOption("ip")),
		},
	}
	ProxyNotFoundError = errors.New("proxy not found")
)

func createFieldFromFilter(filter *pb.Filter) redisearch.Field {
	if filter == nil {
		return nil
	}
	create_prop_filter := func(f *pb.PropertyFilter) redisearch.Field {
		if f == nil {
			return nil
		}
		name := f.GetProperty().GetName()
		value := redisearch.NewStringValue(f.GetValue())
		if f.Op == pb.PropertyFilter_EQUAL {
			return redisearch.NewTagField(name, value)
		} else if f.Op == pb.PropertyFilter_NOT_EQUAL {
			value := redisearch.NewValueOperatorNot(value)
			return redisearch.NewTagField(name, value)
		} else if f.Op == pb.PropertyFilter_NOT_IN {
			value := redisearch.NewValueOperatorNot(value)
			return redisearch.NewTagField(name, value)
		} else if f.Op == pb.PropertyFilter_IN {
			return redisearch.NewTagField(name, value)
		}
		return nil
	}
	create_compo_filter := func(f *pb.CompositeFilter) redisearch.Field {
		if f == nil || f.Filters == nil || len(f.Filters) < 1 {
			return nil
		}
		var field redisearch.Field
		for _, filter := range f.Filters {
			next_field := createFieldFromFilter(filter)
			if next_field == nil {
				continue
			}
			if field == nil {
				field = createFieldFromFilter(filter)
				continue
			}
			if f.GetOp() == pb.CompositeFilter_AND {
				field = redisearch.NewFieldOperatorAnd(field, next_field)
			} else if f.GetOp() == pb.CompositeFilter_OR {
				field = redisearch.NewFieldOperatorOr(field, next_field)
			}
		}
		return field
	}

	switch filter.GetFilterType().(type) {
	case *pb.Filter_PropertyFilter:
		return create_prop_filter(filter.GetPropertyFilter())
	case *pb.Filter_CompositeFilter:
		return create_compo_filter(filter.GetCompositeFilter())
	default:
		return nil
	}
}

const sep = ":"
const proxy_prefix = "proxy"

type ProxyStore struct {
	Proxy
	client *redis.Client
	option StoreOption
	logger *log.Entry
}

func (s *ProxyStore) initIndex() {
	schemas := s.option.Schemas
	if len(schemas) < 1 {
		return
	}
	ctx := context.Background()
	idx_name := "idx:proxy"
	drop_idx := redisearch.FtDropIndex(idx_name)
	create_idx := redisearch.FtCreate(idx_name, redisearch.FTCREATE_ON_JSON, []string{proxy_prefix + sep}, schemas)
	ft_config := redisearch.FtConfigSet("MAXSEARCHRESULTS", "-1")
	pipe := s.client.TxPipeline()
	//TODO: elegant delete index
	ft_config_cmd := pipe.Do(ctx, ft_config...)
	drop_cmd := pipe.Do(ctx, drop_idx...)
	create_cmd := pipe.Do(ctx, create_idx...)
	pipe.Exec(ctx)
	if ft_config_cmd.Err() != nil {
		s.logger.WithField("error", ft_config_cmd.Err()).Error("failed to set config")
	}
	if drop_cmd.Err() != nil {
		s.logger.WithField("error", drop_cmd.Err()).Error("failed to drop index")
		return
	}
	if create_cmd.Err() != nil {
		s.logger.WithField("error", create_cmd.Err()).Error("failed to create index")
	}
}

func (s *ProxyStore) init() {
	s.initIndex()
}

func (s ProxyStore) GetByKey(ctx context.Context, key string, proxy *model.Proxy) error {
	logger := s.logger.WithFields(log.Fields{
		"method": "GetByKey",
	})
	logger.Debugf(key)
	result, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	_proxy := Proxy(*proxy)
	ptr_proxy := &_proxy
	err = ptr_proxy.UnmarshalBinary([]byte(result))
	if err != nil {
		return err
	}
	ret_proxy := model.Proxy(_proxy)
	*proxy = ret_proxy
	return nil
}

func (s ProxyStore) GetByIp(ctx context.Context, ip string, proxy *model.Proxy) error {
	logger := s.logger.WithFields(log.Fields{
		"method": "GetByIp",
		"param":  fmt.Sprintf("%+v", map[string]string{"ip": ip}),
	})
	q := redisearch.NewQuery(1, 0, redisearch.NewTagField("ip", redisearch.NewStringValue(ip)))
	logger.Debug(q)
	search := redisearch.FtSearch("idx:proxy", q)
	logger.Debug(search)
	result, err := s.client.Do(ctx, search...).Result()
	if err != nil {
		logger.WithField("error", err).Error("failed to get proxy")
		return errors.WithStack(err)
	}
	search_result, err := redisearch.ParseSearchResult[Proxy](result)
	if err != nil {
		logger.WithField("error", err).Error("failed to get proxy")
		return errors.WithStack(err)
	}
	if search_result.Total < 1 {
		return errors.WithStack(ProxyNotFoundError)
	}
	*proxy = model.Proxy(search_result.Items[0].Item)
	logger.WithField("result", proxy).Info()
	return nil
}

func (s ProxyStore) GetById(ctx context.Context, id string, proxy *model.Proxy) error {
	logger := s.logger.WithFields(log.Fields{
		"method": "GetById",
		"id":     id,
	})
	q := redisearch.NewQuery(1, 0, redisearch.NewTagField("id", redisearch.NewStringValue(id)))
	logger.Debug(q)
	search := redisearch.FtSearch("idx:proxy", q)
	logger.Debug(search)
	result, err := s.client.Do(ctx, search...).Result()
	if err != nil {
		logger.WithField("error", err).Error("failed to get proxy")
		return errors.WithStack(err)
	}
	search_result, err := redisearch.ParseSearchResult[Proxy](result)
	if err != nil {
		logger.WithField("error", err).Error("failed to get proxy")
		return errors.WithStack(err)
	}
	if search_result.Total < 1 {
		return errors.WithStack(ProxyNotFoundError)
	}
	*proxy = model.Proxy(search_result.Items[0].Item)
	logger.WithField("result", proxy).Info()
	return nil
}

func (s ProxyStore) ExistsId(ctx context.Context, id string) (bool, error) {
	logger := s.logger.WithFields(log.Fields{
		"method": "exists_id",
		"param":  fmt.Sprintf("%+v", map[string]string{"id": id}),
	})
	var _p model.Proxy
	err := s.GetById(ctx, id, &_p)
	if err != nil {
		if errors.Is(err, ProxyNotFoundError) {
			return false, nil
		}
		logger.WithField("error", err).Error()
		return false, err
	}
	logger.WithField("result", true).Info()
	return true, nil
}

func (s ProxyStore) ExistsIp(ctx context.Context, ip string) (bool, error) {
	logger := s.logger.WithFields(log.Fields{
		"method": "exists_ip",
		"param":  fmt.Sprintf("%+v", map[string]string{"ip": ip}),
	})
	var _p model.Proxy
	err := s.GetByIp(ctx, ip, &_p)
	if err != nil {
		if errors.Is(err, ProxyNotFoundError) {
			return false, nil
		}
		logger.WithField("error", err).Error()
		return false, err
	}
	logger.WithField("result", true).Info()
	return true, nil
}

func (s ProxyStore) Add(ctx context.Context, proxy *model.Proxy, options ...SetOption) (*string, error) {
	var setOption SetOption
	if len(options) > 0 {
		setOption = options[0]
	} else {
		setOption = DefaultSetOption
	}
	logger := s.logger.WithFields(log.Fields{
		"method": "Add",
	})
	port_str := fmt.Sprintf("%d", proxy.Port)
	proxy_id, err := common.GenerateUidByStrs(proxy.ProviderId, proxy.ApiId, proxy.Ip, port_str)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	proxy.Id = *proxy_id
	pipe := s.client.TxPipeline()
	proxy_key := strings.Join([]string{proxy_prefix, proxy.Ip, port_str}, sep)
	p := Proxy(*proxy)
	var (
		add_cmd    *redis.StatusCmd
		expire_cmd *redis.BoolCmd
		event_cmd  *redis.StringCmd
	)
	add_cmd = pipe.JSONSet(ctx, proxy_key, "$", p)
	if p.Ttl != -1 {
		expire_cmd = pipe.Expire(ctx, proxy_key, time.Duration(p.Ttl)*time.Second)
	}
	logger.WithFields(
		log.Fields{
			"key":   proxy_key,
			"value": fmt.Sprintf("%#v", proxy),
		},
	).Debug("redis set")
	var stream event.Event
	if setOption.Operation == SETOP_CRETATE {
		stream = event.EVENT_PROXY_CREATED
	} else if setOption.Operation == SETOP_UPDATE {
		stream = event.EVENT_PROXY_UPDATED
	}
	event_cmd = pipe.XAdd(ctx, &redis.XAddArgs{
		Stream:     stream,
		NoMkStream: false,
		Values: map[string]interface{}{
			"proxy": &p,
		}})
	logger.WithFields(
		log.Fields{
			"event": stream,
			"value": fmt.Sprintf("%#v", proxy),
		},
	).Info("event sent")
	pipe.Exec(ctx)
	if add_cmd != nil && add_cmd.Err() != nil {
		err := fmt.Errorf("failed to add proxy %s (error: %+v)", p.Id, add_cmd.Err())
		logger.WithField("error", event_cmd.Err()).Error(err)
		return nil, errors.WithStack(err)
	} else if expire_cmd != nil && expire_cmd.Err() != nil {
		err := fmt.Errorf("failed to set expire for proxy %s (error: %+v)", p.Id, expire_cmd.Err())
		logger.WithField("error", event_cmd.Err()).Error(err)
		return nil, errors.WithStack(err)
	} else if event_cmd != nil && event_cmd.Err() != nil {
		logger.WithField("error", event_cmd.Err()).Error(fmt.Sprintf("failed to send event to stream %s", stream))
	}
	return &proxy.Id, nil
}

func (s ProxyStore) Update(ctx context.Context, id string, proxy model.Proxy, paths []string) error {
	logger := s.logger.WithFields(log.Fields{
		"method": "Update",
		"param":  fmt.Sprintf("%+v", map[string]string{"id": id}),
	})
	var old_proxy model.Proxy
	err := s.GetById(ctx, id, &old_proxy)
	if err != nil {
		return errors.WithStack(err)
	}
	port_str := fmt.Sprintf("%d", old_proxy.Port)
	proxy_key := strings.Join([]string{proxy_prefix, old_proxy.Ip, port_str}, sep)
	p_bytes, _ := json.Marshal(proxy)
	var p_json map[string]interface{}
	_ = json.Unmarshal(p_bytes, &p_json)
	pipe := s.client.TxPipeline()
	//p_str := string(p_bytes)
	//logger.Info(p_str)
	for _, path := range paths {
		path_value, ok := p_json[path]
		if !ok {
			continue
		}
		json_path := fmt.Sprintf("$.%s", path)
		path_value_bytes, _ := json.Marshal(path_value)
		pipe.JSONMerge(ctx, proxy_key, json_path, string(path_value_bytes))
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		err := fmt.Errorf("failed to update proxy %s (err: %+v) ", id, err)
		logger.WithField("error", err).Error(err)
		return errors.WithStack(err)
	}
	return nil
}

func (s ProxyStore) ListWithFilters(ctx context.Context, pager *common.Paginator[model.Proxy], filter *pb.Filter) error {
	logger := s.logger.WithFields(log.Fields{
		"method": "ListWithFilters",
		"param":  fmt.Sprintf("%+v", map[string]string{"filter": fmt.Sprintf("%+v", filter), "offset": fmt.Sprintf("%d", pager.Offset), "limit": fmt.Sprintf("%d", pager.Limit)}),
	})
	logger.Info()
	query_field := createFieldFromFilter(filter)
	logger.Debug(query_field)
	if query_field == nil {
		query_field = redisearch.NewAnyField()
	}
	q := redisearch.NewQuery(int(pager.Limit), int(pager.Offset), query_field)
	logger.Debug(q)
	search := redisearch.FtSearch("idx:proxy", q)
	logger.Debug(search)
	pipe := s.client.TxPipeline()
	logger.Debug(search)
	cmd := pipe.Do(ctx, search...)
	pipe.Exec(ctx)
	result, err := cmd.Result()
	if err != nil {
		logger.WithField("error", cmd.Err()).Error("failed to get proxy")
		return errors.WithStack(cmd.Err())
	}
	search_result, err := redisearch.ParseSearchResult[Proxy](result)
	if err != nil {
		logger.Error(err)
	}
	var ret_proxies []model.Proxy
	for _, v := range search_result.Items {
		ret_proxies = append(ret_proxies, model.Proxy(v.Item))
	}
	pager.Total = int64(search_result.Total)
	pager.Count = int64(len(ret_proxies))
	pager.Items = ret_proxies
	logger.Debugf("%+v", pager)
	return nil
}

func NewProxyStore(client *redis.Client, option ...StoreOption) *ProxyStore {
	var _option StoreOption
	if len(option) == 0 {
		_option = DefaultStoreOption
	} else {
		_option = option[0]
	}
	logger := log.WithFields(
		log.Fields{
			"class": "ProxyStore",
		})
	store := &ProxyStore{client: client, option: _option, logger: logger}
	store.init()
	return store
}
