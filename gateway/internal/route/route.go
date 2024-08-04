package route

import (
	"fmt"

	meta "github.com/WALL-EEEEEEE/proxy-service/gateway/internal/meta"
	"github.com/WALL-EEEEEEE/proxy-service/gateway/log"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/sirupsen/logrus"
)

const default_max_retry = 3

type RouteError struct {
	from string
	to   string
	err  error
}

func (err RouteError) Error() string {
	err_info := fmt.Sprintf("fail to route from %s to %s (err: %+v)", err.from, err.to, err.err.Error())
	return err_info
}

func NewRouteError(from string, to string, err error) RouteError {
	return RouteError{from: from, to: to, err: err}
}

type RouteCallback func(*model.Proxy) error

type RouteOptions struct {
	fallback  *RouteCallback
	metadata  *meta.Metadata
	max_retry *int
}

type RouteOption func(*RouteOptions)

func FallbackRouteOption(fallback RouteCallback) RouteOption {
	return func(options *RouteOptions) {
		options.fallback = &fallback
	}
}
func MetadataRouteOption(metadata meta.Metadata) RouteOption {
	return func(options *RouteOptions) {
		options.metadata = &metadata
	}
}

type RouteRuleMatch[T any] func(T) bool

type RouteRule struct {
	match RouteRuleMatch[any]
}

func NewRouteRule(match RouteRuleMatch[any]) *RouteRule {
	return &RouteRule{match: match}
}

func (r RouteRule) Match(opts ...RouteOption) bool {
	return r.match(opts)
}

type Route[T any] struct {
	rules []RouteRule
	v     T
}

func NewRoute[T any](v T, rules ...RouteRule) *Route[T] {
	return &Route[T]{rules: rules, v: v}
}

func (r Route[T]) Match(opts ...RouteOption) *T {
	options := &RouteOptions{}
	for _, opt := range opts {
		opt(options)
	}
	var matched bool = false
	for _, v := range r.rules {
		if v.Match(opts...) {
			matched = true
		}
	}
	if matched {
		return &r.v
	}
	return nil

}

type RouteTableLoader[T any] ChainLoader[Route[T]]

type RouteTableOptions[T any] struct {
	name        *string
	load_factor *float64
	logger      *log.Logger
	loader      *RouteTableLoader[T]
}

type RouteTableOption[T any] func(*RouteTableOptions[T])

func NameRouteTableOption[T any](name string) RouteTableOption[T] {
	return func(options *RouteTableOptions[T]) {
		options.name = &name
	}
}

func LoadFactorRouteTableOption[T any](factor float64) RouteTableOption[T] {
	return func(options *RouteTableOptions[T]) {
		options.load_factor = &factor
	}
}
func LoaderRouteTableOption[T any](loader RouteTableLoader[T]) RouteTableOption[T] {
	return func(options *RouteTableOptions[T]) {
		options.loader = &loader
	}
}

func LogRouteTableOption[T any](logger *log.Logger) RouteTableOption[T] {
	return func(options *RouteTableOptions[T]) {
		options.logger = logger
	}
}

type RouteTable[T any] struct {
	chain  *EChain[Route[T]]
	logger log.Logger
}

func NewRouteTable[T any](size int, cap int, opts ...RouteTableOption[T]) *RouteTable[T] {
	options := &RouteTableOptions[T]{}
	for _, opt := range opts {
		opt(options)
	}
	var (
		logger log.Logger
	)
	if options.logger != nil {
		logger = *options.logger
	} else {
		logger = log.DefaultLogger
	}
	chain := NewEChain(int32(size), int32(cap), LogEChainOption[Route[T]](&logger), LoadFactorEChainOption[Route[T]](*options.load_factor), LoaderEChainOption(ChainLoader[Route[T]](*options.loader)), NameEChainOption[Route[T]](*options.name))
	var r *RouteTable[T] = &RouteTable[T]{chain: chain, logger: logger}
	return r
}
func (r *RouteTable[T]) Size() int {
	return int(r.chain.Size())
}

func (r *RouteTable[T]) Add(v T) {
	route := NewRoute(v, *NewRouteRule(func(v any) bool { return true }))
	r.chain.Add(*route)
}

func (r *RouteTable[T]) Route(opts ...RouteOption) *T {
	logger := r.logger.WithFields(logrus.Fields{
		"class":  fmt.Sprintf("RouteTable (%p)", r),
		"method": "Route",
	})
	options := &RouteOptions{}
	for _, opt := range opts {
		opt(options)
	}
	for {
		route := r.chain.Next()
		if route == nil {
			return nil
		}
		if (*route).Match(opts...) != nil {
			logger.Debugf("matched: %+v", *route)
			return &(*route).v
		}
	}
}
