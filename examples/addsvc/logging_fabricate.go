package addsvc

import (
	"context"
	"time"

	log "github.com/go-kit/kit/log"
)

type serviceLoggingMiddleware struct {
	logger log.Logger
	next   Service
}

// ServiceLoggingMiddleware returns a service middleware that logs the
// parameters and result of each method invocation.
func ServiceLoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return serviceLoggingMiddleware{
			logger: logger,
			next:   next,
		}
	}
}

// Sum implements Service
func (mw serviceLoggingMiddleware) Sum(ctx context.Context, a int, b int) (result int, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "Sum", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Sum(ctx, a, b)
}

// Concat implements Service
func (mw serviceLoggingMiddleware) Concat(ctx context.Context, a string, b string) (result string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "Concat", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Concat(ctx, a, b)
}
