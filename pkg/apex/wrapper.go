// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package apex wraps apex/log
package apex

import (
	"github.com/TheThingsIndustries/mystique/pkg/log"
	apex "github.com/apex/log"
)

type wrapper struct {
	apex.Interface
}

func (w wrapper) WithField(k string, v interface{}) log.Interface {
	return &wrapper{w.Interface.WithField(k, v)}
}
func (w wrapper) WithFields(fields log.Fielder) log.Interface {
	return &wrapper{w.Interface.WithFields(apex.Fields(fields.Fields()))}
}
func (w wrapper) WithError(err error) log.Interface {
	return &wrapper{w.Interface.WithError(err)}
}

// Wrap the apex logger
func Wrap(w apex.Interface) log.Interface {
	return &wrapper{w}
}

// Log is the global apex logger
var Log log.Interface = &wrapper{apex.Log}

// SetLevelFromString proxies to the global apex logger
func SetLevelFromString(s string) { apex.SetLevelFromString(s) }
