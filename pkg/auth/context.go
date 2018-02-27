// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package auth

import "context"

type ctxKeyType struct{}

var ctxKey ctxKeyType

// InterfaceFromContext returns the auth interface from the context
func InterfaceFromContext(ctx context.Context) Interface {
	if v := ctx.Value(ctxKey); v != nil {
		if auth, ok := v.(Interface); ok {
			return auth
		}
	}
	return nil
}

// NewContextWithInterface returns a new context that contains the interface
func NewContextWithInterface(ctx context.Context, auth Interface) context.Context {
	return context.WithValue(ctx, ctxKey, auth)
}
