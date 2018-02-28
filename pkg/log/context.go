// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package log

import (
	"context"
)

type ctxKeyType struct{}

var ctxKey ctxKeyType

// FromContext calls log.FromContext
func FromContext(ctx context.Context) Interface {
	if v := ctx.Value(ctxKey); v != nil {
		if logger, ok := v.(Interface); ok {
			return logger
		}
	}
	return Noop
}

// NewContext returns a new context that contains the logger
func NewContext(ctx context.Context, logger Interface) context.Context {
	return context.WithValue(ctx, ctxKey, logger)
}
