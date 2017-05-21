package addsvc

import (
	context "context"
	fmt "fmt"
	time "time"

	endpoint "github.com/go-kit/kit/endpoint"
	log "github.com/go-kit/kit/log"
	metrics "github.com/go-kit/kit/metrics"
)

type Endpoints struct {
	SumEndpoint    endpoint.Endpoint
	ConcatEndpoint endpoint.Endpoint
}

type sumRequest struct {
	A int
	B int
}

type sumResponse struct {
	Result int
	Err    error
}

func (e Endpoints) Sum(ctx context.Context, a int, b int) (result int, err error) {
	request := sumRequest{
		A: a,
		B: b,
	}
	response, err := e.SumEndpoint(ctx, request)
	if err != nil {
		return 0, err
	}
	return response.(sumResponse).Result, response.(sumResponse).Err
}

func MakeSumEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		sumReq := request.(sumRequest)
		result, err := s.Sum(ctx, sumReq.A, sumReq.B)
		return sumResponse{
			Err:    err,
			Result: result,
		}, nil
	}
}

type concatRequest struct {
	A string
	B string
}

type concatResponse struct {
	Result string
	Err    error
}

func (e Endpoints) Concat(ctx context.Context, a string, b string) (result string, err error) {
	request := concatRequest{
		A: a,
		B: b,
	}
	response, err := e.ConcatEndpoint(ctx, request)
	if err != nil {
		return "", err
	}
	return response.(concatResponse).Result, response.(concatResponse).Err
}

func MakeConcatEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		concatReq := request.(concatRequest)
		result, err := s.Concat(ctx, concatReq.A, concatReq.B)
		return concatResponse{
			Err:    err,
			Result: result,
		}, nil
	}
}

func EndpointInstrumentingMiddleware(duration metrics.Histogram) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				duration.With("success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
			}(time.Now())

			return next(ctx, request)
		}
	}
}

func EndpointLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				logger.Log("error", err, "took", time.Since(begin))
			}(time.Now())

			return next(ctx, request)
		}
	}
}
