package job

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/provider-adapter/param"

	manager "github.com/WALL-EEEEEEE/proxy-service/manager"
	event "github.com/WALL-EEEEEEE/proxy-service/manager/event"
	pb "github.com/WALL-EEEEEEE/proxy-service/manager/gen/manager/v1"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/WALL-EEEEEEE/proxy-service/manager/util"

	retry_http "github.com/hashicorp/go-retryablehttp"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
)

type ProxyCheck struct {
	proxy  model.Proxy
	client *http.Client
}

func NewProxyCheck(proxy model.Proxy) ProxyCheck {
	retryClient := retry_http.NewClient()
	retryClient.RetryMax = 10
	return ProxyCheck{proxy: proxy, client: retryClient.StandardClient()}
}

func (p *ProxyCheck) Check() {
	assess := manager.NewAssess()
	log.Debugf("check proxy : %s", p.proxy.Ip)
	assess.Run(&p.proxy)
	stats := manager.Collect(&p.proxy, true)
	//add protocol support
	/*
		var support_proto []model.PROTO
		if stats.HttpSupport {
			support_proto = append(support_proto, model.PROTO_HTTP)
		}
		if stats.SocketSupport {
			support_proto = append(support_proto, model.PROTO_SOCKET)
		}
		if stats.WebSocketSupport {
			support_proto = append(support_proto, model.PROTO_WEBSOCKET)
		}
		p.proxy.Proto = support_proto
	*/
	if p.proxy.Attr == nil {
		p.proxy.Attr = &model.Attr{}
	}
	if stats.Dialable && stats.Latency > 0 {
		p.proxy.Attr.Availiable = true
	}
	p.proxy.Attr.Latency = stats.Latency
	now := time.Now()
	p.proxy.CheckedAt = &now
	p.proxy.Status = model.STATUS_CHECKED
	log.Infof("%+v", p.proxy)
}

func (p ProxyCheck) GetCron() string {
	return ""
}

func (p ProxyCheck) GetId() string {
	return p.proxy.Id
}

func (p ProxyCheck) GetName() string {
	return p.proxy.Ip
}

func (p ProxyCheck) GetDesc() string {
	return "Check proxy " + p.proxy.Ip
}

type ProxyCheckJob struct {
	event_bus             *redis.Client
	event_group_id        string
	event_begin_id        string
	id                    string
	manager_client        pb.ProxyServiceClient
	proxy_channel         chan param.RawProxy
	check_channel         chan CheckItem
	checked_proxy_channel chan model.Proxy
	checker_quota         int
	ctx                   context.Context
	cancel                context.CancelFunc
	stop                  chan bool
}

func NewProxyCheckJob(ctx context.Context, id string, manager_api string, event_bus *redis.Client) (*ProxyCheckJob, error) {
	ctx, cancel := context.WithCancel(ctx)
	proxy_channel := make(chan param.RawProxy)
	check_channel := make(chan CheckItem)
	checked_proxy_channel := make(chan model.Proxy)
	log.Infof("start proxy check job ...")
	conn, err := grpc.Dial(manager_api, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	manager_client := pb.NewProxyServiceClient(conn)
	job := &ProxyCheckJob{ctx: ctx, checker_quota: 5, manager_client: manager_client, event_bus: event_bus, id: id, event_group_id: "proxy_check_job", event_begin_id: ">", proxy_channel: proxy_channel, check_channel: check_channel, checked_proxy_channel: checked_proxy_channel, cancel: cancel, stop: make(chan bool)}
	return job, nil
}

func (p *ProxyCheckJob) startProxyEventsListener() {
	logger := log.WithFields(log.Fields{
		"task": "proxy_event_listener",
	})
	logger.Info("start")
	var resp *redis.StatusCmd
	events := []event.Event{
		event.EVENT_PROXY_CREATED,
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
						logger.Debugf("msg: %+v (id: %s)", msg.Values, msg_id)
						check_data, ok := msg.Values["proxy"].(string)
						if !ok {
							logger.Warnf("invalid message format: %+v (message must contains proxy field)", msg.Values)
							continue
						}
						var proxy model.Proxy
						err := json.Unmarshal([]byte(check_data), &proxy)
						if err != nil {
							logger.Warnf("invalid message format: %s", err.Error())
							continue
						}
						logger.Infof("%+v", proxy)
						check_op := NewProxyCheck(proxy)
						p.AddCheck(&check_op)
					}
				}

				switch event_type {
				case event.EVENT_PROXY_CREATED:
					process_created_messages(event_msgs)
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

func (p *ProxyCheckJob) startProxyChecker() {
	var wg sync.WaitGroup
	check_func := func(channel chan CheckItem) {
		defer wg.Done()
		for {
			check := <-p.check_channel
			check.Check()
			p.checked_proxy_channel <- check.(*ProxyCheck).proxy
		}
	}
	for i := 0; i < p.checker_quota; i++ {
		wg.Add(1)
		go check_func(p.check_channel)
	}
	wg.Wait()
}

func (p *ProxyCheckJob) startLoadBuzzer() {
	logger := log.WithFields(log.Fields{
		"task": "load_buzzer",
	})

	stevedore_load := func(proxy model.Proxy) error {
		logger.Infof("proxy: %+v", proxy)
		now := time.Now()
		proxy.UpdatedAt = &now
		field_mask, err := fieldmaskpb.New(&pb.Proxy{}, []string{"proto", "updated_at", "checked_at", "status", "attr"}...)
		if err != nil {
			return err
		}
		req := &pb.UpdateProxyRequest{
			Id:     proxy.Id,
			Proxy:  util.PbFromProxy(&proxy),
			Fields: field_mask,
		}
		logger.Info(req.String())
		resp, err := p.manager_client.UpdateProxy(p.ctx, req)
		if err != nil {
			return err
		}
		if resp.Status.Code != 0 {
			return fmt.Errorf("error while load proxy: %s", resp.Status.Message)
		}
		logger.Infof("%s loaded ", proxy.Ip)
		return nil
	}
	for {
		proxy := <-p.checked_proxy_channel
		err := stevedore_load(proxy)
		if err != nil {
			logger.Error(err)
			continue
		}
	}
}

func (p *ProxyCheckJob) Start() {
	go p.startProxyEventsListener()
	go p.startProxyChecker()
	go p.startLoadBuzzer()
}

func (p *ProxyCheckJob) Join() {
	<-p.stop
}

func (p *ProxyCheckJob) Close() {
	p.cancel()
	close(p.stop)
}

func (p *ProxyCheckJob) AddCheck(check CheckItem) {
	p.check_channel <- check
}
