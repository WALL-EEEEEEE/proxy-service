package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/common"

	"github.com/WALL-EEEEEEE/proxy-service/gateway/internal/handler"
	meta "github.com/WALL-EEEEEEE/proxy-service/gateway/internal/meta"
	route "github.com/WALL-EEEEEEE/proxy-service/gateway/internal/route"
	server "github.com/WALL-EEEEEEE/proxy-service/gateway/internal/server"
	"github.com/WALL-EEEEEEE/proxy-service/gateway/internal/util"
	log "github.com/WALL-EEEEEEE/proxy-service/gateway/log"
	model "github.com/WALL-EEEEEEE/proxy-service/manager/model"

	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	CONNECT_TIMEOUT = 10
	DAIL_TIMEOUT    = 10
)

func real_addr(req http.Request) string {
	real_addr := req.Host
	if _, port, _ := net.SplitHostPort(real_addr); port == "" {
		real_addr = net.JoinHostPort(real_addr, "80")
	}
	return real_addr
}
func real_port(req http.Request) int {
	_, port, _ := net.SplitHostPort(req.Host)
	if port == "" {
		return 80
	}
	iport, _ := strconv.Atoi(port)
	return iport
}
func http_proto(req http.Request) string {
	if req.Method == http.MethodConnect {
		return "https"
	} else {
		return "http"
	}
}

func proxy_http_request(ctx context.Context, conn net.Conn, req http.Request, proxy *model.Proxy) error {
	resp := &http.Response{
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	var (
		wrap_addr string
		req_addr  string = real_addr(req)
	)
	if proxy != nil {
		wrap_addr = net.JoinHostPort(proxy.UseConfig.Host, strconv.Itoa(int(proxy.UseConfig.Port)))
	} else {
		wrap_addr = req_addr
	}
	d := net.Dialer{}
	dail_ctx, cancel := context.WithTimeout(ctx, DAIL_TIMEOUT*time.Second)
	defer cancel()
	wrap_conn, err := d.DialContext(dail_ctx, "tcp", wrap_addr)
	if err != nil {
		resp.StatusCode = http.StatusServiceUnavailable
		resp.Body = io.NopCloser(bytes.NewBufferString("proxy unreachable"))
		resp.Write(conn)
		return err
	}
	defer wrap_conn.Close()

	if proxy != nil {
		wrap_req := &http.Request{
			Method:     http.MethodConnect,
			URL:        &url.URL{Host: req_addr},
			Host:       req_addr,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     http.Header{},
		}
		wrap_req.Header.Set("Proxy-Connection", "keep-alive")
		if proxy.UseConfig != nil && proxy.UseConfig.User != "" && proxy.UseConfig.Password != "" {
			u := proxy.UseConfig.User
			p := proxy.UseConfig.Password
			wrap_req.Header.Set("Proxy-Authorization",
				"Basic "+base64.StdEncoding.EncodeToString([]byte(u+":"+p)))
		}
		connect_ctx, cancel := context.WithTimeout(ctx, CONNECT_TIMEOUT*time.Second)
		defer cancel()
		wrap_req = wrap_req.WithContext(connect_ctx)
		if err := wrap_req.Write(wrap_conn); err != nil {
			return err
		}
		wrap_resp, err := http.ReadResponse(bufio.NewReader(wrap_conn), wrap_req)
		if err != nil {
			return err
		}
		if wrap_resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s", wrap_resp.Status)
		}
	}
	if req.Method == http.MethodConnect {
		//proxy through connect protocol
		resp.StatusCode = http.StatusOK
		resp.Status = "200 Connection established"
		if err := resp.Write(conn); err != nil {
			return err
		}
	} else {
		//proxy directly
		req.Header.Del("Proxy-Connection")
		if err := req.Write(wrap_conn); err != nil {
			return err
		}
	}
	err = util.Transport(conn, wrap_conn)
	return err

}

func auto_proxy(ctx context.Context, handler *handler.HttpHandler, conn net.Conn, req *http.Request) (err error) {
	logger := logger.WithFields(
		logrus.Fields{
			"class":  "HttpHandler",
			"handle": "auto_proxy",
		})

	var metadata meta.Metadata = meta.Metadata{}
	var target_addr string = real_addr(*req)

	metadata["addr"] = target_addr
	metadata["proto"] = http_proto(*req)
	if req.Header != nil {
		header_str, _ := json.Marshal(req.Header)
		metadata["header"] = string(header_str)
	}
	cb := func(proxy *model.Proxy) error {
		start := time.Now()
		defer func() {
			logger := logger.WithFields(logrus.Fields{
				"cost": fmt.Sprintf(" %.2fs", time.Since(start).Seconds()),
			})
			if proxy == nil {
				logger.Infof("redirect %s -> %s (direct) ", target_addr, "localhost")
			} else {
				logger.Infof("redirect %s -> %s (proxied) ", target_addr, proxy.Ip)
			}
		}()
		err := proxy_http_request(ctx, conn, *req, proxy)
		if err != nil && proxy != nil {
			return route.NewRouteError(proxy.Ip, target_addr, err)
		}
		return err
	}
	brouter.Route(ctx, cb, route.FallbackRouteOption(cb), route.MetadataRouteOption(metadata))
	return nil
}

var (
	port        int
	manager_api string
	loglevel    string
	logger      *logrus.Logger
	brouter     *route.ProxyBrouter
	cmd         = &cobra.Command{
		Use:   "http",
		Short: "http proxy server",
		Run: func(cmd *cobra.Command, args []string) {
			_logger := log.Logger{Logger: logger}
			var err error
			common.SetLevel(loglevel)
			if err != nil {
				logger.Error(err)
				return
			}
			ctx := context.Background()
			brouter, err = route.NewProxyBrouter(ctx, manager_api, route.LogProxyBrouterOption(&_logger), route.RouteTableCapProxyBrouterOption(1000), route.RouteTableSizeProxyBrouterOption(20))
			if err != nil {
				logger.Error(err)
				return
			}
			http_serv, err := server.NewHttpProxyServer(
				port,
				server.LogHttpProxyServerOption(&_logger),
				server.HandleHttpProxyServerOption(auto_proxy),
			)
			if err != nil {
				logger.Error(err)
				return
			}
			http_serv.Serve()
		},
	}
)

func main() {
	cmd.Flags().IntVarP(&port, "port", "p", 8000, "port listened on")
	cmd.Flags().StringVarP(&manager_api, "manager-api", "m", "", "grpc service address of proxy service")
	cmd.Flags().StringVarP(&loglevel, "log", "l", "INFO", "log level")
	cmd.MarkFlagRequired("manager-api")
	common.SetupLog("class", "method")
	logger = logrus.StandardLogger()
	if err := cmd.Execute(); err != nil {
		cmd.PrintErr(err)
		os.Exit(1)
	}
}
