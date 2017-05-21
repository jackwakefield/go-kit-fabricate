package addsvc

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
)

type serviceInstrumentingMiddleware struct {
	requestCount   metrics.Counter
	errorCount     metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

func ServiceIntrumentingMiddleware(requestCount, errorCount metrics.Counter, requestLatency metrics.Histogram) Middleware {
	return func(next Service) Service {
		return serviceInstrumentingMiddleware{
			errorCount:     errorCount,
			next:           next,
			requestCount:   requestCount,
			requestLatency: requestLatency,
		}
	}
}

// Sum implements Service
func (mw serviceInstrumentingMiddleware) Sum(ctx context.Context, a int, b int) (result int, err error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "Sum").Add(1)
		mw.requestLatency.With("method", "Sum").Observe(time.Since(begin).Seconds())
		if err != nil {
			mw.errorCount.With("method", "Sum").Add(1)
		}
	}(time.Now())
	return mw.next.Sum(ctx, a, b)
}

// Concat implements Service
func (mw serviceInstrumentingMiddleware) Concat(ctx context.Context, a string, b string) (result string, err error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "Concat").Add(1)
		mw.requestLatency.With("method", "Concat").Observe(time.Since(begin).Seconds())
		if err != nil {
			mw.errorCount.With("method", "Concat").Add(1)
		}
	}(time.Now())
	return mw.next.Concat(ctx, a, b)
}
