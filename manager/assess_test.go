package manager

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"

	"github.com/WALL-EEEEEEE/Axiom/test"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/bobg/go-generics/slices"
	"github.com/sirupsen/logrus"
	logrus_test "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockConn struct {
	mock.Mock
	net.Conn
}

func (conn *MockConn) Close() error {
	return nil
}

func TestAssess(t *testing.T) {
	proxies := []Proxy{
		{
			Ip:    "58.20.82.115",
			Port:  2323,
			Proto: []model.PROTO{model.PROTO_HTTP},
		},
	}
	ping_ttls := []int64{
		15,
		20,
		53,
		78,
		160,
	}

	cases := []test.TestCase[any, any]{
		{
			Name:  "Ping.HappyPath",
			Input: proxies,
			Error: nil,
			Expected: func() []Stats {
				var stats_list []Stats
				for _, proxy := range proxies {
					stats := newStats(&proxy)
					stats_list = append(stats_list, stats)
				}
				return stats_list
			},
			Check: func(tc test.TestCase[any, any]) {
				patches := gomonkey.ApplyFunc(ping_timeout, func(proxy string, timeout int, times int) ([]int64, error) {
					return ping_ttls, nil
				})
				defer patches.Reset()
				for _, proxy := range tc.Input.([]Proxy) {
					proxy_assess := NewAssess()
					proxy_assess.Ping(&proxy)
					proxy_assess.Wait()
				}
				proxy_stats := CollectAll(true)
				expect_proxy_stats := tc.Expected.(func() []Stats)()
				assert.Equal(t, expect_proxy_stats, proxy_stats)
			},
		},
		{
			Name:     "Ping.InvalidProxy",
			Input:    proxies,
			Error:    nil,
			Expected: nil,
			Check: func(tc test.TestCase[any, any]) {
				hook := logrus_test.NewGlobal()
				defer hook.Reset()
				patches := gomonkey.ApplyFunc(net.ResolveIPAddr, func(network string, address string) (*net.IPAddr, error) {
					return nil, &net.AddrError{Err: "invalid address", Addr: "xxxxx"}
				})
				defer patches.Reset()
				for _, proxy := range tc.Input.([]Proxy) {
					proxy_assess := NewAssess()
					proxy_assess.Ping(&proxy)
					proxy_assess.Wait()
				}
				expected_error_log := "address xxxxx: invalid address"

				expected_log_entry, _ := slices.Filter[*logrus.Entry](hook.AllEntries(), func(entry *logrus.Entry) (bool, error) {
					return entry.Message == expected_error_log, nil
				})
				require.Equal(t, len(expected_log_entry), 1)
				assert.Equal(t, expected_log_entry[0].Level, logrus.ErrorLevel)
			},
		},
		{
			Name:  "Dial.HappyPath",
			Input: proxies,
			Error: nil,
			Expected: func() []Stats {
				var stats_list []Stats
				for _, proxy := range proxies {
					stats := newStats(&proxy)
					stats.Dialable = true
					stats_list = append(stats_list, stats)
				}
				return stats_list
			},
			Check: func(tc test.TestCase[any, any]) {
				conn := new(MockConn)
				mock_close := conn.On("Close").Return(nil)
				patches := gomonkey.ApplyFunc(dial_timeout, func(network, address string, timeout time.Duration) (net.Conn, error) {
					return conn, nil
				})
				defer patches.Reset()
				defer mock_close.Unset()
				for _, proxy := range tc.Input.([]Proxy) {
					proxy_assess := NewAssess()
					proxy_assess.Dial(&proxy)
					proxy_assess.Wait()
				}
				proxy_stats := CollectAll(true)
				expect_proxy_stats := tc.Expected.(func() []Stats)()
				assert.Equal(t, proxy_stats, expect_proxy_stats)
			},
		},
		{
			Name:  "Dial.Timeout",
			Input: proxies,
			Error: nil,
			Expected: func() []Stats {
				var stats_list []Stats
				for _, proxy := range proxies {
					stats := newStats(&proxy)
					stats.Dialable = false
					stats_list = append(stats_list, stats)
				}
				return stats_list
			},
			Check: func(tc test.TestCase[any, any]) {
				hook := logrus_test.NewGlobal()
				conn := new(MockConn)
				mock_close := conn.On("Close").Return(nil)
				patches := gomonkey.ApplyFunc(net.DialTimeout, func(network, address string, timeout time.Duration) (net.Conn, error) {
					return nil, errors.New("i/o timeout")
				})
				defer patches.Reset()
				defer mock_close.Unset()
				for _, proxy := range tc.Input.([]Proxy) {
					proxy_assess := NewAssess()
					proxy_assess.Dial(&proxy)
					proxy_assess.Wait()
				}
				proxy_stats := CollectAll(true)
				expect_proxy_stats := tc.Expected.(func() []Stats)()
				for i, proxy_stats := range proxy_stats {
					assert.Equal(t, expect_proxy_stats[i], proxy_stats)
				}
				expected_error_log := "i/o timeout"

				expected_log_entry, _ := slices.Filter[*logrus.Entry](hook.AllEntries(), func(entry *logrus.Entry) (bool, error) {
					return entry.Message == expected_error_log, nil
				})
				require.Equal(t, len(expected_log_entry), 1)
				assert.Equal(t, expected_log_entry[0].Level, logrus.ErrorLevel)
			},
		},
		{
			Name:  "Dial.InvalidProxy",
			Input: proxies,
			Error: nil,
			Expected: func() []Stats {
				var stats_list []Stats
				for _, proxy := range proxies {
					stats := newStats(&proxy)
					stats.Dialable = false
					stats_list = append(stats_list, stats)
				}
				return stats_list
			},
			Check: func(tc test.TestCase[any, any]) {
				hook := logrus_test.NewGlobal()
				defer hook.Reset()
				conn := new(MockConn)
				mock_close := conn.On("Close").Return(nil)
				defer mock_close.Unset()
				patches := gomonkey.ApplyFunc(net.DialTimeout, func(network, address string, timeout time.Duration) (net.Conn, error) {
					return nil, &net.AddrError{Err: "invalid address", Addr: "xxxxx"}
				})
				defer patches.Reset()
				for _, proxy := range tc.Input.([]Proxy) {
					proxy_assess := NewAssess()
					proxy_assess.Dial(&proxy)
					proxy_assess.Wait()
				}
				proxy_stats := CollectAll(true)
				expect_proxy_stats := tc.Expected.(func() []Stats)()
				for i, proxy_stats := range proxy_stats {
					assert.Equal(t, expect_proxy_stats[i], proxy_stats)
				}
				expected_error_log := "address xxxxx: invalid address"

				expected_log_entry, _ := slices.Filter[*logrus.Entry](hook.AllEntries(), func(entry *logrus.Entry) (bool, error) {
					return entry.Message == expected_error_log, nil
				})
				require.Equal(t, len(expected_log_entry), 1)
				assert.Equal(t, expected_log_entry[0].Level, logrus.ErrorLevel)
			},
		},
		{
			Name:  "Http.HappyPath",
			Input: proxies,
			Error: nil,
			Expected: func() []Stats {
				var stats_list []Stats
				for _, proxy := range proxies {
					stats := newStats(&proxy)
					stats.HttpSupport = true
					stats_list = append(stats_list, stats)
				}
				return stats_list
			},
			Check: func(tc test.TestCase[any, any]) {
				patches := gomonkey.ApplyFunc(http_test, func(url string, proxy_url string, timeout int) (int, string, error) {
					resp_text := "{\"ip\":\"58.20.82.115\",\"country\":\"CN\",\"asn\":{\"asnum\":4837,\"org_name\":\"CHINA UNICOM China169 Backbone\"},\"geo\":{\"city\":\"\",\"region\":\"\",\"region_name\":\"\",\"postal_code\":\"\",\"latitude\":34.7732,\"longitude\":113.722,\"tz\":\"Asia/Shanghai\"}}"
					return 200, resp_text, nil
				})
				defer patches.Reset()

				for _, proxy := range tc.Input.([]Proxy) {
					proxy_assess := NewAssess()
					proxy_assess.Http(&proxy)
					proxy_assess.Wait()
				}
				proxy_stats := CollectAll(true)
				expect_proxy_stats := tc.Expected.(func() []Stats)()
				assert.Equal(t, expect_proxy_stats, proxy_stats)
			},
		},
		{
			Name:  "WebSocket.HappyPath",
			Input: proxies,
			Error: nil,
			Expected: func() []Stats {
				var stats_list []Stats
				for _, proxy := range proxies {
					stats := newStats(&proxy)
					stats.WebSocketSupport = true
					stats_list = append(stats_list, stats)
				}
				return stats_list
			},
			Check: func(tc test.TestCase[any, any]) {
				patches := gomonkey.ApplyFunc(websocket_test, func(url string, proxy_url string, timeout int) error {
					return nil
				})
				defer patches.Reset()
				for _, proxy := range tc.Input.([]Proxy) {
					proxy_assess := NewAssess()
					proxy_assess.WebSocket(&proxy)
					proxy_assess.Wait()
				}
				proxy_stats := CollectAll(true)
				expect_proxy_stats := tc.Expected.(func() []Stats)()
				assert.Equal(t, proxy_stats, expect_proxy_stats)
			},
		},
		{
			Name:  "Socket.HappyPath",
			Input: proxies,
			Error: nil,
			Expected: func() []Stats {
				var stats_list []Stats
				for _, proxy := range proxies {
					stats := newStats(&proxy)
					stats.SocketSupport = true
					stats_list = append(stats_list, stats)
				}
				return stats_list
			},
			Check: func(tc test.TestCase[any, any]) {
				patches := gomonkey.ApplyFunc(socket_test, func(url string, proxy_url string, timeout int) error {
					return nil
				})
				defer patches.Reset()
				for _, proxy := range tc.Input.([]Proxy) {
					proxy_assess := NewAssess()
					proxy_assess.Socket(&proxy)
					proxy_assess.Wait()
				}
				proxy_stats := CollectAll(true)
				expect_proxy_stats := tc.Expected.(func() []Stats)()
				assert.Equal(t, proxy_stats, expect_proxy_stats)
			},
		},

		{
			Name:  "Run.HappyPath",
			Input: proxies,
			Error: nil,
			Expected: func() []Stats {
				var stats_list []Stats
				for _, proxy := range proxies {
					stats := newStats(&proxy)
					stats.HttpSupport = true
					stats.Dialable = true
					stats.WebSocketSupport = true
					stats.SocketSupport = true
					stats_list = append(stats_list, stats)
				}
				return stats_list
			},
			Check: func(tc test.TestCase[any, any]) {
				conn := new(MockConn)
				mock_close := conn.On("Close").Return(nil)
				defer mock_close.Unset()
				patches := gomonkey.ApplyFunc(http_test, func(url string, proxy_url string, timeout int) (int, string, error) {
					resp_text := "{\"ip\":\"58.20.82.115\",\"country\":\"CN\",\"asn\":{\"asnum\":4837,\"org_name\":\"CHINA UNICOM China169 Backbone\"},\"geo\":{\"city\":\"\",\"region\":\"\",\"region_name\":\"\",\"postal_code\":\"\",\"latitude\":34.7732,\"longitude\":113.722,\"tz\":\"Asia/Shanghai\"}}"
					return 200, resp_text, nil
				})
				patches = patches.ApplyFunc(net.DialTimeout, func(network, address string, timeout time.Duration) (net.Conn, error) {
					return conn, nil
				})
				patches = patches.ApplyFunc(ping_timeout, func(proxy string, timeout int, times int) ([]int64, error) {
					return ping_ttls, nil
				})
				patches = patches.ApplyFunc(websocket_test, func(url string, proxy_url string, timeout int) error {
					return nil
				})
				patches = patches.ApplyFunc(socket_test, func(url string, proxy_url string, timeout int) error {
					return nil
				})
				defer patches.Reset()
				for _, proxy := range tc.Input.([]Proxy) {
					proxy_assess := NewAssess()
					proxy_assess.Run(&proxy)
				}
				proxy_stats := CollectAll(true)
				expect_proxy_stats := tc.Expected.(func() []Stats)()
				assert.Equal(t, proxy_stats, expect_proxy_stats)
			},
		},
	}
	test.Run(cases, t)
}
