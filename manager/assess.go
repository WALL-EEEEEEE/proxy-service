package manager

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	net_url "net/url"
	"strings"
	"sync"
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"

	"github.com/bobg/go-generics/maps"
	"github.com/gorilla/websocket"
	cache "github.com/patrickmn/go-cache"
	ping "github.com/prometheus-community/pro-bing"
	log "github.com/sirupsen/logrus"
	net_proxy "golang.org/x/net/proxy"
)

var dial_timeout = net.DialTimeout

type Proxy = model.Proxy

type Params struct {
	PingMaxRTT       int
	DialTimeOut      int
	HttpTimeOut      int
	PingTimes        int
	WebSocketTimeOut int
	SocketTimeOut    int
}

var DefaultParams Params = Params{
	PingMaxRTT:       1,
	PingTimes:        5,
	DialTimeOut:      10,
	HttpTimeOut:      15,
	WebSocketTimeOut: 15,
	SocketTimeOut:    15,
}

var http_api = "https://lumtest.com/myip.json"
var websocket_api = "wss://ws.postman-echo.com/raw/"
var socket_api = "tcpbin.com:4242"
var statsCache = cache.New(cache.NoExpiration, cache.NoExpiration)
var mutStatsUpdate sync.Mutex

type Stats struct {
	proxy            *Proxy
	Latency          int64
	Dialable         bool
	HttpSupport      bool
	WebSocketSupport bool
	SocketSupport    bool
}

type UpdateStatsOption func(stats Stats) Stats

var newStats = func(proxy *Proxy) Stats {
	return Stats{
		proxy:       proxy,
		Latency:     0,
		Dialable:    false,
		HttpSupport: false,
	}
}

var getStats = func(proxy *Proxy) *Stats {
	ip_str := fmt.Sprintf("%s,%s,%d", proxy.Provider, proxy.Ip, proxy.Port)
	mutStatsUpdate.Lock()
	cache_stats, _ := statsCache.Get(ip_str)
	mutStatsUpdate.Unlock()
	_cache_stats := cache_stats.(Stats)
	return &_cache_stats
}

var deleteStats = func(proxy *Proxy) {
	ip_str := fmt.Sprintf("%s,%s,%d", proxy.Provider, proxy.Ip, proxy.Port)
	mutStatsUpdate.Lock()
	statsCache.Delete(ip_str)
	mutStatsUpdate.Unlock()
}

var updateStats = func(proxy *Proxy, update_option UpdateStatsOption) {
	var stats Stats
	ip_str := fmt.Sprintf("%s,%s,%d", proxy.Provider, proxy.Ip, proxy.Port)
	mutStatsUpdate.Lock()
	cache_stats, exist := statsCache.Get(ip_str)
	if !exist {
		stats = newStats(proxy)
	} else {
		stats = cache_stats.(Stats)
	}
	stats = update_option(stats)
	statsCache.Set(ip_str, stats, cache.NoExpiration)
	mutStatsUpdate.Unlock()
}

var ping_timeout = func(ip string, timeout int, times int) ([]int64, error) {
	var ping_rtts []int64
	p, err := ping.NewPinger(ip)
	if err != nil {
		return nil, err
	}
	p.Count = times
	p.Timeout = time.Second * time.Duration(timeout*times)
	p.OnRecv = func(p *ping.Packet) {
		ping_rtts = append(ping_rtts, p.Rtt.Milliseconds())
	}
	p.SetPrivileged(true)
	err = p.Run()
	if err != nil {
		return nil, err
	}
	return ping_rtts, nil
}

type Assess struct {
	wg    sync.WaitGroup
	param Params
}

func NewAssess(param ...Params) Assess {
	_param := DefaultParams
	if len(param) > 1 {
		_param = param[0]
	}
	return Assess{
		param: _param,
	}
}

func http_test(url string, proxy string, timeout int) (int, string, error) {
	_proxy, err := net_url.Parse(proxy)
	if err != nil {
		return 0, "", err
	}
	log.Debugf("Http Get %s via proxy %+v", url, _proxy)
	tr := &http.Transport{
		Proxy:           http.ProxyURL(_proxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * time.Duration(timeout), //超时时间
	}
	resp, err := client.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, string(body), err
	}
	return resp.StatusCode, string(body), nil
}

func websocket_test(url string, proxy string, timeout int) error {
	websocket.DefaultDialer.Proxy = func(req *http.Request) (*net_url.URL, error) {
		return net_url.Parse(proxy)
	}
	websocket.DefaultDialer.HandshakeTimeout = time.Second * time.Duration(timeout)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer c.Close()
	return nil
}
func socket_test(url string, proxy string, timeout int) error {
	var network = "tcp"
	dialer, err := net_proxy.SOCKS5(network, proxy, nil, net_proxy.Direct)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()
	c, err := dialer.(net_proxy.ContextDialer).DialContext(ctx, network, url)
	if err != nil {
		return err
	}
	defer c.Close()
	return nil
}

func (assess *Assess) Ping(proxy *Proxy) {
	assess.wg.Add(1)
	logger := log.WithFields(log.Fields{
		"task":      "assess",
		"metric":    "ping",
		"max_rtt":   assess.param.PingMaxRTT,
		"max_times": assess.param.PingTimes,
		"ip":        proxy.Ip,
	})
	logger.Info("start")
	go func() {
		var ping_durations []int64
		defer assess.wg.Done()
		defer func(begin time.Time) {
			updateStats(proxy, func(stats Stats) Stats {
				stats.Latency = func([]int64) int64 {
					var ttl int64
					for _, duration := range ping_durations {
						ttl += duration
					}
					return ttl / int64(len(ping_durations))
				}(ping_durations)
				return stats
			})
			logger.WithField("cost", fmt.Sprintf(" %fs", time.Since(begin).Seconds())).Infof("%v", ping_durations)
		}(time.Now())
		ping_ip := proxy.UseConfig.Host
		ping_durations, err := ping_timeout(ping_ip, assess.param.PingMaxRTT, assess.param.PingTimes)
		if err != nil {
			logger.Error(err)
		}
	}()
}
func (assess *Assess) Wait() {
	assess.wg.Wait()
}

func (assess *Assess) Dial(proxy *Proxy) {
	assess.wg.Add(1)
	logger := log.WithFields(log.Fields{
		"task":    "assess",
		"metric":  "dail",
		"timeout": fmt.Sprintf(" %ds", assess.param.DialTimeOut),
		"ip":      proxy.Ip,
	})
	logger.Info("start")
	go func() {
		defer assess.wg.Done()
		var dailable bool
		defer func(begin time.Time) {
			updateStats(proxy, func(stats Stats) Stats {
				stats.Dialable = dailable
				return stats
			})
			logger.WithField("cost", fmt.Sprintf(" %fs", time.Since(begin).Seconds())).Infof("%v", dailable)
		}(time.Now())
		dail_host := fmt.Sprintf("%s:%d", proxy.UseConfig.Host, proxy.UseConfig.Port)
		conn, err := dial_timeout("tcp", dail_host, time.Second*time.Duration(assess.param.DialTimeOut))
		if conn != nil {
			defer conn.Close()
		}
		if err != nil {
			log.Error(err)
			dailable = false
			return
		}
		dailable = true

	}()

}

func (assess *Assess) Http(proxy *Proxy) {
	assess.wg.Add(1)
	logger := log.WithFields(log.Fields{
		"task":    "assess",
		"metric":  "http",
		"timeout": fmt.Sprintf(" %ds", assess.param.HttpTimeOut),
		"ip":      proxy.Ip,
	})
	logger.Info("start")
	go func() {
		var http_support bool
		defer assess.wg.Done()
		defer func(begin time.Time) {
			updateStats(proxy, func(stats Stats) Stats {
				stats.HttpSupport = http_support
				return stats
			})
			logger.WithField("cost", fmt.Sprintf(" %fs", time.Since(begin).Seconds())).Infof("%v", http_support)
		}(time.Now())
		proxy_url := proxy.UseConfig.Psn
		code, resp, err := http_test(http_api, proxy_url, assess.param.HttpTimeOut)
		if err != nil {
			log.Error(err)
			http_support = false
			return
		}
		if code == 200 && strings.Contains(resp, proxy.Ip) {
			http_support = true
		} else {
			http_support = false
		}
		log.Debugf(resp)
	}()
}
func (assess *Assess) Socket(proxy *Proxy) {
	assess.wg.Add(1)
	logger := log.WithFields(log.Fields{
		"task":    "assess",
		"metric":  "socket",
		"timeout": fmt.Sprintf(" %ds", assess.param.SocketTimeOut),
		"ip":      proxy.Ip,
	})
	logger.Info("start")
	go func() {
		var socket_support bool
		defer assess.wg.Done()
		defer func(begin time.Time) {
			updateStats(proxy, func(stats Stats) Stats {
				stats.SocketSupport = socket_support
				return stats
			})
			logger.WithField("cost", fmt.Sprintf(" %ds", time.Since(begin))).Infof("%v", socket_support)
		}(time.Now())
		proxy_url := proxy.UseConfig.Psn
		err := socket_test(socket_api, proxy_url, assess.param.SocketTimeOut)
		if err != nil {
			log.Error(err)
			socket_support = false
			return
		}
		socket_support = true
	}()

}
func (assess *Assess) WebSocket(proxy *Proxy) {
	assess.wg.Add(1)
	logger := log.WithFields(log.Fields{
		"task":    "assess",
		"metric":  "websocket",
		"timeout": fmt.Sprintf(" %ds", assess.param.WebSocketTimeOut),
		"ip":      proxy.Ip,
	})
	logger.Info("start")
	go func() {
		var websocket_support bool
		defer assess.wg.Done()
		defer func(begin time.Time) {
			updateStats(proxy, func(stats Stats) Stats {
				stats.WebSocketSupport = websocket_support
				return stats
			})
			logger.WithField("cost", fmt.Sprintf(" %fs", time.Since(begin).Seconds())).Infof("%v", websocket_support)
		}(time.Now())
		proxy_url := proxy.UseConfig.Psn
		err := websocket_test(websocket_api, proxy_url, assess.param.WebSocketTimeOut)
		if err != nil {
			log.Error(err)
			websocket_support = false
			return
		}
		websocket_support = true
	}()

}
func (assess *Assess) Concurrency(p *Proxy) {}
func (assess *Assess) Speed(p *Proxy)       {}
func (assess *Assess) Stability(p *Proxy)   {}

func Collect(proxy *Proxy, delete bool) *Stats {
	stats := getStats(proxy)
	if stats != nil && delete {
		deleteStats(proxy)
	}
	return stats
}

func CollectAll(delete bool) []Stats {
	var all_stats []Stats
	maps.Each(statsCache.Items(), func(key string, item cache.Item) error {
		all_stats = append(all_stats, item.Object.(Stats))
		return nil
	})
	if delete {
		statsCache.Flush()
	}
	return all_stats
}
func (assess *Assess) Run(proxy *Proxy) {
	logger := log.WithFields(log.Fields{
		"task": "assess",
		"ip":   proxy.Ip,
	})
	defer func(begin time.Time) {
		logger.WithField("cost", fmt.Sprintf(" %fs", time.Since(begin).Seconds())).Infof("%+v", Collect(proxy, false))
	}(time.Now())
	logger.Info("start")
	assess.Ping(proxy)
	assess.Dial(proxy)
	support_protos := proxy.Proto
	for _, proto := range support_protos {
		switch proto {
		case model.PROTO_HTTP:
			assess.Http(proxy)
		case model.PROTO_SOCKET:
			assess.Socket(proxy)
		case model.PROTO_WEBSOCKET:
			assess.WebSocket(proxy)
		}
	}
	assess.Wait()
}
