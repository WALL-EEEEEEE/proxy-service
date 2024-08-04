package route

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"github.com/WALL-EEEEEEE/proxy-service/gateway/log"
	"github.com/sirupsen/logrus"
)

const default_load_factor = 0.5

// TODO: optimize the chain implementation
type Chain[T any] interface {
	Next() *T
	Value() *T
}

type ChainLoader[T any] func(size int) []T

type EChainOptions[T any] struct {
	logger      *log.Logger
	ctx         *context.Context
	loader      *ChainLoader[T]
	load_factor *float64
	name        *string
}

type EChainOption[T any] func(*EChainOptions[T])

func NameEChainOption[T any](name string) EChainOption[T] {
	return func(options *EChainOptions[T]) {
		options.name = &name
	}
}

func LogEChainOption[T any](logger *log.Logger) EChainOption[T] {
	return func(options *EChainOptions[T]) {
		options.logger = logger
	}
}
func CtxEChainOption[T any](ctx *context.Context) EChainOption[T] {
	return func(options *EChainOptions[T]) {
		options.ctx = ctx
	}
}
func LoadFactorEChainOption[T any](factor float64) EChainOption[T] {
	return func(options *EChainOptions[T]) {
		options.load_factor = &factor
	}
}
func LoaderEChainOption[T any](loader ChainLoader[T]) EChainOption[T] {
	return func(options *EChainOptions[T]) {
		options.loader = &loader
	}
}

type EChain[T any] struct {
	name        string
	start       int32
	end         int32
	max_end     int32
	res         int32
	init_cap    int32
	cntr        *[]T
	load_factor float64
	logger      log.Logger
	loader      ChainLoader[T]
}

func NewEChain[T any](size, cap int32, opts ...EChainOption[T]) *EChain[T] {
	options := &EChainOptions[T]{}
	for _, opt := range opts {
		opt(options)
	}
	cntr := make([]T, 0, cap)
	chain := &EChain[T]{cntr: &cntr, start: 0, end: 0, res: 0, init_cap: size}
	if options.logger != nil {
		chain.logger = *options.logger
	} else {
		chain.logger = log.DefaultLogger
	}
	if options.load_factor != nil {
		chain.load_factor = *options.load_factor
	} else {
		chain.load_factor = default_load_factor
	}
	if options.name != nil {
		chain.name = *options.name
	}
	chain.loader = *options.loader
	*(chain.cntr) = (*(chain.cntr))[:size]
	go sync.OnceFunc(func() { chain.prefetch(int(size)) })()
	return chain
}

func (c *EChain[T]) Size() int {
	return len(*c.cntr)
}

func (c *EChain[T]) Cap() int {
	return cap(*c.cntr)
}

func (c *EChain[T]) Add(vs ...T) {
	logger := c.logger.WithFields(logrus.Fields{
		"class":  fmt.Sprintf("Echain-%s (%p)", c.name, c),
		"method": "Add",
	})
	for _, v := range vs {
		(*c.cntr)[c.end] = v
		atomic.CompareAndSwapInt32(&c.end, c.end, (c.end+1)%int32(c.Cap()))
		atomic.CompareAndSwapInt32(&c.res, c.res, (c.res+1)%int32(c.Cap()))
		atomic.CompareAndSwapInt32(&c.max_end, c.max_end, int32(math.Max(float64(c.max_end), float64(c.end))))
		logger.WithFields(logrus.Fields{
			"size":  c.Size(),
			"cap":   c.Cap(),
			"res":   c.res,
			"start": c.start,
			"end":   c.end,
		}).Debugf("%+v", v)
	}
}

func (c *EChain[T]) Next() *T {
	logger := c.logger.WithFields(logrus.Fields{
		"class":  fmt.Sprintf("Echain-%s (%p)", c.name, c),
		"method": "Next",
	})
	if c.cntr == nil || len(*c.cntr) == 0 || c.max_end == 0 {
		go sync.OnceFunc(func() { c.prefetch(int(c.init_cap)) })()
		return nil
	}
	val := (*c.cntr)[c.start]
	atomic.CompareAndSwapInt32(&c.start, c.start, (c.start+1)%c.max_end)
	if c.res-1 < 0 {
		atomic.CompareAndSwapInt32(&c.res, c.res, 0)
	} else {
		atomic.CompareAndSwapInt32(&c.res, c.res, c.res-1)
	}
	go sync.OnceFunc(c.grow)()
	logger.WithFields(logrus.Fields{
		"size":  c.Size(),
		"cap":   c.Cap(),
		"res":   c.res,
		"start": c.start,
		"end":   c.end,
	}).Debugf("%+v", val)
	return &val
}

func (c *EChain[T]) Value() *T {
	if c.cntr == nil || len(*c.cntr) == 0 {
		return nil
	}
	val := (*c.cntr)[c.start]
	return &val
}

func (c *EChain[T]) prefetch(size int) {
	cnt := 0
	vs := c.loader(size)
	for _, v := range vs {
		c.Add(v)
		cnt++
	}
}

func (c *EChain[T]) grow() {
	logger := c.logger.WithFields(logrus.Fields{
		"class":  fmt.Sprintf("EChain-%s (%p)", c.name, c),
		"method": "grow",
	})
	factor := 1 - float64(c.res)/float64(c.Size()) // curent factor
	logger = logger.WithFields(
		logrus.Fields{
			"start":       c.start,
			"end":         c.end,
			"size":        c.Size(),
			"cap":         c.Cap(),
			"res":         c.res,
			"factor":      factor,
			"load_factor": c.load_factor,
		},
	)
	if factor < c.load_factor {
		logger.Debugf("no need to expand")
	} else {
		last_res := c.res
		t_size := len(*c.cntr)>>1 + len(*c.cntr) //target size
		e_size := t_size - len(*c.cntr)
		next_cntr_size := int32(math.Min(float64(c.Cap()), float64(t_size)))
		*(c.cntr) = (*(c.cntr))[:next_cntr_size]
		if e_size > math.MaxInt32 {
			e_size = math.MaxInt32
		}
		c.prefetch(e_size)
		logger.WithFields(logrus.Fields{
			"start":  c.start,
			"end":    c.end,
			"res":    c.res,
			"target": t_size,
		}).Debugf("%d -> %d (+%d)", last_res, c.res, c.res-last_res)
	}
}
