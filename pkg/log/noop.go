// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package log

// Noop logger
var Noop Interface = &noop{}

type noop struct{}

func (n *noop) Debug(msg string)                        {}
func (n *noop) Info(msg string)                         {}
func (n *noop) Warn(msg string)                         {}
func (n *noop) Error(msg string)                        {}
func (n *noop) Fatal(msg string)                        {}
func (n *noop) Debugf(msg string, v ...interface{})     {}
func (n *noop) Infof(msg string, v ...interface{})      {}
func (n *noop) Warnf(msg string, v ...interface{})      {}
func (n *noop) Errorf(msg string, v ...interface{})     {}
func (n *noop) Fatalf(msg string, v ...interface{})     {}
func (n *noop) WithField(string, interface{}) Interface { return n }
func (n *noop) WithFields(Fielder) Interface            { return n }
func (n *noop) WithError(error) Interface               { return n }
