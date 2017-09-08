// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package log defines the MQTT log interface.
package log

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// Fielder alias
type Fielder = log.Fielder

// Interface alias
type Interface = log.Interface

// Fields calls log.Fields
func Fields(pairs ...interface{}) Fielder { return log.Fields(pairs...) }

// FromContext calls log.FromContext
func FromContext(ctx context.Context) Interface { return log.Ensure(log.FromContext(ctx)) }
