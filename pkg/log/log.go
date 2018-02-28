// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package log defines the MQTT log interface.
package log

// Fielder is the interface for anything that can have fields.
type Fielder interface {
	Fields() map[string]interface{}
}

// Interface is the interface for logging.
type Interface interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	Debugf(msg string, v ...interface{})
	Infof(msg string, v ...interface{})
	Warnf(msg string, v ...interface{})
	Errorf(msg string, v ...interface{})
	Fatalf(msg string, v ...interface{})
	WithField(string, interface{}) Interface
	WithFields(Fielder) Interface
	WithError(error) Interface
}

type F map[string]interface{}

func (f F) Fields() map[string]interface{} {
	return f
}

// Fields returns a Fielder
func Fields(f map[string]interface{}) Fielder { return F(f) }
