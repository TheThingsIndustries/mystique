// Copyright © 2021 The Things Industries, distributed under the MIT license (see LICENSE file)

package ratelimit

import (
	"context"

	"golang.org/x/time/rate"
)

type ctxKeyType struct{}

var ctxKey ctxKeyType

// New returns a new rate limiter that allows limit events per second.
func New(parent context.Context, limit rate.Limit) context.Context {
	limiter := rate.NewLimiter(limit, 1)
	return context.WithValue(parent, ctxKey, limiter)
}

// Wait blocks until the rate limit configuration permits an event to happen.
// It returns an error if the Context is canceled, or the expected wait time
// exceeds the Context's Deadline.
func Wait(ctx context.Context) (err error) {
	if limiter, ok := ctx.Value(ctxKey).(*rate.Limiter); ok {
		return limiter.Wait(ctx)
	}
	return nil
}
