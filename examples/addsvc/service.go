package addsvc

//go:generate fabricate $GOFILE

// This file contains the Service definition, and a basic service
// implementation. It also includes service middlewares.

import (
	"context"
	"errors"
)

// Service describes a service that adds things together.
type Service interface {
	Sum(ctx context.Context, a, b int) (result int, err error)
	Concat(ctx context.Context, a, b string) (result string, err error)
}

// Business-domain errors like these may be served in two ways: returned
// directly by endpoints, or bundled into the response struct. Both methods can
// be made to work, but errors returned directly by endpoints are counted by
// middlewares that check errors, like circuit breakers.
//
// If you don't want that behavior -- and you probably don't -- then it's better
// to bundle errors into the response struct.

var (
	// ErrTwoZeroes is an arbitrary business rule for the Add method.
	ErrTwoZeroes = errors.New("can't sum two zeroes")

	// ErrIntOverflow protects the Add method. We've decided that this error
	// indicates a misbehaving service and should count against e.g. circuit
	// breakers. So, we return it directly in endpoints, to illustrate the
	// difference. In a real service, this probably wouldn't be the case.
	ErrIntOverflow = errors.New("integer overflow")

	// ErrMaxSizeExceeded protects the Concat method.
	ErrMaxSizeExceeded = errors.New("result exceeds maximum size")
)

// NewBasicService returns a naïve, stateless implementation of Service.
func NewBasicService() Service {
	return basicService{}
}

type basicService struct{}

const (
	intMax = 1<<31 - 1
	intMin = -(intMax + 1)
	maxLen = 102400
)

// Sum implements Service.
func (s basicService) Sum(_ context.Context, a, b int) (int, error) {
	if a == 0 && b == 0 {
		return 0, ErrTwoZeroes
	}
	if (b > 0 && a > (intMax-b)) || (b < 0 && a < (intMin-b)) {
		return 0, ErrIntOverflow
	}
	return a + b, nil
}

// Concat implements Service.
func (s basicService) Concat(_ context.Context, a, b string) (string, error) {
	if len(a)+len(b) > maxLen {
		return "", ErrMaxSizeExceeded
	}
	return a + b, nil
}
