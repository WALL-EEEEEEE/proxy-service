package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	event "github.com/WALL-EEEEEEE/proxy-service/manager/event"
	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/WALL-EEEEEEE/proxy-service/manager/repository"
	"github.com/WALL-EEEEEEE/proxy-service/manager/util"

	common_param "github.com/WALL-EEEEEEE/proxy-service/common/param"

	adapter_param "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/param"

	"strings"

	"github.com/bobg/go-generics/maps"
	retry_http "github.com/hashicorp/go-retryablehttp"
	"github.com/redis/go-redis/v9"
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type RawProxy = adapter_param.RawProxy

type CheckItem interface {
	Check()
	GetName() string
	GetDesc() string
	GetCron() string
	GetId() string
}

const default_limit = 100

type InvalidArgumentError struct {
	msg string
}

func (e InvalidArgumentError) Error() string {
	return e.msg
}

type UnavailableError struct {
	msg string
}

func (e UnavailableError) Error() string {
	return e.msg
}

type ProxyApiCheck struct {
	proxy_api     model.ProxyApi
	client        *http.Client
	proxy_channel chan<- ProxyCheckItem
	batch_process chan<- string
	db            *gorm.DB
	cache         map[string]interface{}
}

type ProxyCheckItem struct {
	Proxy   model.Proxy
	CheckId string
}

func NewProxyApiCheck(proxy_api model.ProxyApi, proxy_channel chan<- ProxyCheckItem, batch_process chan<- string, store *gorm.DB) ProxyApiCheck {
	retryClient := retry_http.NewClient()
	retryClient.RetryMax = 0
	client := retryClient.StandardClient()
	cache := make(map[string]interface{})
	return ProxyApiCheck{proxy_api: proxy_api, client: client, proxy_channel: proxy_channel, batch_process: batch_process, db: store, cache: cache}
}

func (p ProxyApiCheck) proxy_adapter_handler(client *http.Client, adapter_api string, params map[string]string, results chan<- ProxyCheckItem) {
	provider_info, _ := p.cache["provider"].(repository.ProxyProvider)
	api_info, _ := p.cache["api"].(repository.ProxyApi)

	request_func := func(req adapter_param.ListProxiesRequest) (*adapter_param.ListProxiesResponse, error) {
		req_data, err := json.Marshal(req)
		if err != nil {
			return nil, InvalidArgumentError{msg: err.Error()}
		}
		resp, err := client.Post(adapter_api, "application/json", bytes.NewBuffer(req_data))
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			if err.(*url.Error).Op == "parse" {
				err = InvalidArgumentError{msg: err.Error()}
			}
			err = UnavailableError{msg: err.Error()}
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, UnavailableError{msg: fmt.Sprintf("invalid response (status: %s)", resp.Status)}
		}
		var resp_content []byte
		resp_content, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, UnavailableError{msg: fmt.Sprintf("invalid response (error: %s, status: %s, response: %.100s)", err.Error(), resp.Status, resp_content)}
		}
		var api_adatper_resp adapter_param.ListProxiesResponse
		err = json.Unmarshal(resp_content, &api_adatper_resp)
		if err != nil {
			return nil, UnavailableError{msg: fmt.Sprintf("invalid response (error: %s, status: %s, response: %.100s)", err.Error(), resp.Status, resp_content)}
		}
		if api_adatper_resp.Status.Code != common_param.STATUS_OK.Code {
			return nil, UnavailableError{msg: fmt.Sprintf("invalid response (status: %s, response: %.100s)", resp.Status, resp_content)}
		}
		return &api_adatper_resp, nil
	}
	go func() {
		var limit, offset int64 = default_limit, 0
		for {
			log.Infof("get proxy from proxy adapter %s (limit: %d, offset: %d) ...", adapter_api, limit, offset)
			adapter_req := adapter_param.ListProxiesRequest{Pager: common_param.Pager{Limit: limit, Offset: offset}, Params: params}
			adapter_resp, err := request_func(adapter_req)
			if err != nil {
				log.Error(err)
				return
			}
			for _, proxy := range adapter_resp.Proxies {
				now := time.Now()
				_proxy := model.Proxy{
					ProviderId: p.proxy_api.ProviderId,
					ApiId:      p.proxy_api.Id,
					Api:        api_info.Name,
					Provider:   provider_info.Name,
					Ip:         proxy.Ip,
					Port:       proxy.Port,
					Ttl:        proxy.Ttl,
					Proto:      proxy.Proto,
					Attr:       proxy.Attr,
					UseConfig:  proxy.UseConfig,
					Status:     model.STATUS_CREATED,
					CreatedAt:  &now,
				}
				results <- ProxyCheckItem{Proxy: _proxy, CheckId: p.GetId()}
			}
			offset = adapter_resp.Page.Offset
			log.Debugf("Offset: %d", offset)
			if offset >= adapter_resp.Page.Total || len(adapter_resp.Proxies) >= int(adapter_resp.Page.Total) {
				p.batch_process <- p.GetId()
				break
			}
		}
	}()
}
func (p ProxyApiCheck) get_extra_info() error {
	_, ok := p.cache["api"]
	if !ok {
		var api repository.ProxyApi
		result := p.db.Where("api_id = ?", p.proxy_api.Id).First(&api)
		if result.Error != nil {
			return fmt.Errorf("failed to query api from db (error: %+v)", result.Error)
		}
		p.cache["api"] = api
	}
	_, ok = p.cache["provider"]
	if !ok {
		var provider repository.ProxyProvider
		result := p.db.Where("provider_id = ?", p.proxy_api.ProviderId).First(&provider)
		if result.Error != nil {
			return fmt.Errorf("failed to query provider from db (error: %+v)", result.Error)
		}
		p.cache["provider"] = provider
	}
	return nil
}

func (p ProxyApiCheck) Check() {
	//construct api url from the api specified
	err := p.get_extra_info()
	if err != nil {
		log.Error(err)
		return
	}
	api_name := p.proxy_api.Service.Name
	api_host := p.proxy_api.Service.Host
	api_url := strings.Join([]string{api_host, strings.ReplaceAll(api_name, ".", "/")}, "/")
	api_params := p.proxy_api.Service.Params
	p.proxy_adapter_handler(p.client, api_url, api_params, p.proxy_channel)
}

func (p ProxyApiCheck) GetId() string {
	return p.proxy_api.Id
}

func (p ProxyApiCheck) GetName() string {
	return p.proxy_api.Name
}

func (p ProxyApiCheck) GetCron() string {
	cron_expr := fmt.Sprintf("@every %ds", int(p.proxy_api.UpdateInterval))
	return cron_expr
}

func (p ProxyApiCheck) GetDesc() string {
	return "Check proxy api " + p.proxy_api.Name
}

type ProxyApiCheckJob struct {
	executor       *cron.Cron
	checks         map[string]CheckItem
	entry_to_check map[cron.EntryID]CheckItem
	check_to_entry map[string]cron.EntryID
	event_bus      *redis.Client
	event_group_id string
	event_begin_id string
	id             string
	db             *gorm.DB
	proxy_service  pb.ProxyServiceClient
	channel        chan ProxyCheckItem
	batch_process  chan string
	ctx            context.Context
	cancel         context.CancelFunc
	stop           chan bool
}

func NewProxyApiCheckJob(ctx context.Context, id string, manager_api string, event_bus *redis.Client, db *gorm.DB) (*ProxyApiCheckJob, error) {
	grpc_conn, err := grpc.Dial(manager_api, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	channel := make(chan ProxyCheckItem)
	entry_to_check := make(map[cron.EntryID]CheckItem)
	check_to_entry := make(map[string]cron.EntryID)
	checks := make(map[string]CheckItem)
	proxy_client := pb.NewProxyServiceClient(grpc_conn)
	inst := &ProxyApiCheckJob{
		ctx:            ctx,
		executor:       cron.New(cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger))),
		event_bus:      event_bus,
		db:             db,
		id:             id,
		event_group_id: "proxy_api_check_job",
		event_begin_id: ">",
		channel:        channel,
		batch_process:  make(chan string),
		cancel:         cancel,
		stop:           make(chan bool),
		entry_to_check: entry_to_check,
		check_to_entry: check_to_entry,
		checks:         checks,
		proxy_service:  proxy_client,
	}
	return inst, nil
}

func (p *ProxyApiCheckJob) createProxy(logger log.Entry, proxy model.Proxy) error {
	logger.Debugf("proxy: %+v", proxy)
	*proxy.CreatedAt = time.Now()

	var add_proxy = util.PbFromProxy(&proxy)
	req := &pb.AddProxyRequest{
		Proxy:      add_proxy,
		ProviderId: proxy.ProviderId,
		ApiId:      proxy.ApiId,
	}
	resp, err := p.proxy_service.AddProxy(p.ctx, req)
	if err != nil {
		return err
	}
	if resp.Status.Code != 0 {
		return fmt.Errorf("error while load proxy: %s", resp.Status.Message)
	}
	logger.Debugf("loaded in %s", resp.Id)
	return nil
}

func (p *ProxyApiCheckJob) startLoadBuzzer() {
	logger := log.WithFields(log.Fields{
		"task": "load_buzzer",
	})
	logger.Infof("start")
	var load_status map[string]int64 = make(map[string]int64)
	for {
		select {
		case proxy_check_item := <-p.channel:
			err := p.createProxy(*logger, proxy_check_item.Proxy)
			if err != nil {
				errStatus, ok := status.FromError(err)
				if !(ok && errStatus.Code() == codes.AlreadyExists) {
					logger.Error(err)
				}
				continue
			}
			load_status[proxy_check_item.CheckId] += 1
		case check_id := <-p.batch_process:
			check_load_status := load_status[check_id]
			logger.Infof("check %s loaded %d proxy", check_id, check_load_status)
		}

	}
}

func (p *ProxyApiCheckJob) startCheckEventsListener() {
	logger := log.WithFields(log.Fields{
		"task": "check_event_listener",
	})
	logger.Info("start")
	var resp *redis.StatusCmd
	events := []event.Event{
		event.EVENT_CHECKAPI_CREATED,
		event.EVENT_CHECKAPI_UPDATED,
		event.EVENT_CHECKAPI_DELETED,
	}
	for _, event := range events {
		resp = p.event_bus.XGroupCreateMkStream(p.ctx, event, p.event_group_id, "0")
		if resp.Err() != nil && resp.Err().Error() != "BUSYGROUP Consumer Group name already exists" {
			logger.Errorf("error when creating consumer group [%s] in stream [%s] : %s", p.event_group_id, event, resp.Err())
			return
		}
	}
	logger.Infof("subscribe to event streams: %+v", events)
	for {
		select {
		case <-p.ctx.Done():
			logger.Info("exit")
			return
		default:
			var streams []string = append(events, strings.Split(strings.Repeat(">", len(events)), "")...)
			results, err := p.event_bus.XReadGroup(p.ctx, &redis.XReadGroupArgs{Group: p.event_group_id, Consumer: p.id, Streams: streams}).Result()
			if err != nil {
				logger.Errorf("error while reading event from event bus: %s", err)
				continue
			}
			process_event := func(event_type event.Event, event_msgs []redis.XMessage) error {
				logger = logger.WithFields(log.Fields{
					"event": event_type,
				})
				process_created_messages := func(msgs []redis.XMessage) {
					for _, msg := range msgs {
						msg_id := msg.ID
						logger.Infof("msg: %+v (id: %s)", msg.Values, msg_id)
						check_data, ok := msg.Values["data"].(string)
						if !ok {
							logger.Warnf("invalid message format: %+v (message must contains data field)", msg.Values)
							continue
						}
						var proxy_api model.ProxyApi
						err := json.Unmarshal([]byte(check_data), &proxy_api)
						if err != nil {
							logger.Warnf("invalid message format: %s", err.Error())
							continue
						}
						check_op := NewProxyApiCheck(proxy_api, p.channel, p.batch_process, p.db)
						p.AddCheck(check_op)
					}
				}
				process_updated_messages := func(msgs []redis.XMessage) {
					for _, msg := range msgs {
						msg_id := msg.ID
						logger.Infof("msg: %+v (id: %s)", msg.Values, msg_id)
						check_data, ok := msg.Values["data"].(string)
						if !ok {
							logger.Warnf("invalid message format: %+v (message must contains data field)", msg.Values)
							continue
						}
						var proxy_api model.ProxyApi
						err := json.Unmarshal([]byte(check_data), &proxy_api)
						if err != nil {
							logger.Warnf("invalid message format: %s", err.Error())
							continue
						}
						check_op := NewProxyApiCheck(proxy_api, p.channel, p.batch_process, p.db)
						p.RemoveCheck(check_op.GetName())
						p.AddCheck(check_op)
					}
				}
				process_deleted_message := func(msgs []redis.XMessage) {
					for _, msg := range msgs {
						msg_id := msg.ID
						logger.Infof("msg: %+v (id: %s)", msg.Values, msg_id)
						check_data, ok := msg.Values["data"].(string)
						if !ok {
							logger.Warnf("invalid message format: %+v (message must contains 'data' field)", msg.Values)
							continue
						}
						var proxy_api model.ProxyApi
						err := json.Unmarshal([]byte(check_data), &proxy_api)
						if err != nil {
							logger.Warnf("invalid message format: %s", err.Error())
							continue
						}
						check_op := NewProxyApiCheck(proxy_api, p.channel, p.batch_process, p.db)
						p.RemoveCheck(check_op.GetName())
					}
				}

				switch event_type {
				case event.EVENT_CHECKAPI_CREATED:
					process_created_messages(event_msgs)
				case event.EVENT_CHECKAPI_UPDATED:
					process_updated_messages(event_msgs)
				case event.EVENT_CHECKAPI_DELETED:
					process_deleted_message(event_msgs)
				default:
					logger.Warnf("unknown event")
				}
				return nil
			}
			for _, stream := range results {
				err := process_event(stream.Stream, stream.Messages)
				if err != nil {
					logger.Error(err)
				}
			}
		}
	}
}

func (p *ProxyApiCheckJob) startCheckSynchronizer() {
	logger := log.WithFields(log.Fields{
		"task": "api_check_sync",
	})

	logger.Infof("start")

	p.executor.AddFunc("@every 1m", func() {
		logger.Debugf("retrieve api items in database ...")
		var apis []repository.ProxyApi
		result := p.db.Find(&apis)
		if result.Error != nil {
			logger.Errorf("failed to query api checks from db (error: %+v)", result.Error)
			return
		}
		logger.Debugf("found %d apis in db, sync it into tasks", len(apis))
		var add_cnt, remove_cnt int
		var mutual_checks map[string]interface{} = make(map[string]interface{})
		for _, api := range apis {
			logger.Debugf("API: %+v", api)
			check := NewProxyApiCheck(model.ProxyApi{
				Service: model.Service{
					Host:   api.Service.Host,
					Name:   api.Service.Name,
					Params: api.Service.Params,
				},
				UpdateInterval: api.UpdateInterval,
				Name:           api.Name,
				ProviderId:     api.ProviderId,
				Id:             api.ApiId,
			}, p.channel, p.batch_process, p.db)
			mutual_checks[check.GetId()] = struct{}{}
			_, ok := p.checks[check.GetName()]
			if !ok {
				logger.Debugf("check %s not found, try to add it ...", check.GetName())
				p.AddCheck(check)
				add_cnt += 1
				continue
			}
		}
		maps.Each(p.checks, func(key string, value CheckItem) error {
			_, ok := mutual_checks[value.GetId()]
			if !ok {
				logger.Infof("check %s not found, try to remove it ...", value.GetName())
				p.RemoveCheck(value.GetName())
				remove_cnt += 1
			}
			return nil
		})
		logger.Infof("synced db tasks (add: %d, remove: %d)", add_cnt, remove_cnt)
	})
}

func (p *ProxyApiCheckJob) Start() {
	go p.startLoadBuzzer()
	go p.startCheckEventsListener()
	go p.startCheckSynchronizer()
	p.executor.Start()
}

func (p *ProxyApiCheckJob) Join() {
	<-p.stop
}

func (p *ProxyApiCheckJob) Close() {
	p.executor.Stop()
	p.cancel()
	close(p.stop)
}

func (p *ProxyApiCheckJob) AddCheck(check CheckItem) error {
	log.Infof("Add check %s (ID:%s, Desc: %s, Cron: %s)...", check.GetName(), check.GetId(), check.GetDesc(), check.GetCron())
	entry_id, err := p.executor.AddFunc(check.GetCron(), func() {
		log.Infof("Checking for %s (ID:%s, Desc: %s, Cron: %s) ...", check.GetName(), check.GetId(), check.GetDesc(), check.GetCron())
		check.Check()
	})
	if err != nil {
		return err
	}
	p.entry_to_check[entry_id] = check
	p.check_to_entry[check.GetId()] = entry_id
	p.checks[check.GetName()] = check
	return nil
}

func (p *ProxyApiCheckJob) RemoveCheck(check_name string) error {
	check, exist := p.checks[check_name]
	if !exist {
		return fmt.Errorf("check %s not exists", check_name)
	}
	entry_id := p.check_to_entry[check.GetId()]
	log.Infof("Remove check %s ...", check.GetName())
	p.executor.Remove(entry_id)
	delete(p.entry_to_check, entry_id)
	delete(p.check_to_entry, check.GetId())
	delete(p.checks, check.GetName())
	return nil
}
