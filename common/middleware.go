package common

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/go-kit/kit/endpoint"
)

// AppendKeyvalser is an interface that wraps the basic AppendKeyvals method.
//
// AppendKeyvals should be implemented to append key/value pairs into keyvals
// without removing any existing elements, then return the extended keyvals.
//
//	Example:
//		// Define your struct type
//		type SomeType struct{
//			AField string
//			BField string
//		}
//
//		// Implement the AppendKeyvals func to satisfy the AppendKeyvalser interface
//		func (s SomeType) AppendKeyvals(keyvals []interface{}) []interface{} {
//			// Add key/value sets here (2 values per set, key followed by value)
//		 	return append(keyvals,
//		 		"SomeType.AField", s.AField,
//		 		"SomeType.BField", s.BField)
//		}
type AppendKeyvalser interface {
	AppendKeyvals(keyvals []interface{}) []interface{}
}

const (
	tookKey     = "took"
	transErrKey = "transport_error"
)

// LoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, the resulting error (if any), and
// keyvals specific to the request and response object if they implement
// the AppendKeyvalser interface. If not nil, errLogger will be used for logging
// requests that resulted in a non-nil error being returned.
func LoggingMiddleware(logger, errLogger *log.Logger) endpoint.Middleware {
	if logger == nil {
		return nil
	}
	if errLogger == nil {
		errLogger = logger
	}
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				kvs := makeKeyvals(request, response, time.Since(begin), err)
				if err != nil {
					errLogger.WithFields(kvs).Error()
				} else {
					logger.WithFields(kvs).Info()
				}
			}(time.Now())
			return next(ctx, request)
		}
	}
}

type StackTracer interface {
	StackTrace() errors.StackTrace
}

// makeKeyvals will place the received parameters into an []interface{} to be
// returned in the order:
//  1. err
//  2. d
//  3. req (if AppendKeyvalser is implemented)
//  4. resp (if AppendKeyvalser is implemented)
func makeKeyvals(req, resp interface{}, d time.Duration, err error) log.Fields {
	KVs := log.Fields{
		transErrKey: err,
		tookKey:     fmt.Sprintf("%.2f s", d.Seconds()),
	}
	if err != nil {
		err, ok := err.(StackTracer)
		if ok {
			KVs["stacktrace"] = fmt.Sprintf("%+v", err)
		}
	}
	if l, ok := req.(AppendKeyvalser); ok {
		_kvs := []interface{}{}
		_kvs = l.AppendKeyvals(_kvs)
		for i := 0; i < len(_kvs); i += 2 {
			key := _kvs[i].(string)
			value := fmt.Sprintf("%+v", _kvs[i+1])
			KVs[key] = value
		}
	}
	if l, ok := resp.(AppendKeyvalser); ok {
		_kvs := []interface{}{}
		_kvs = l.AppendKeyvals(_kvs)
		for i := 0; i < len(_kvs); i += 2 {
			key := _kvs[i].(string)
			value := fmt.Sprintf("%+v", _kvs[i+1])
			KVs[key] = value
		}
	}
	return KVs
}
