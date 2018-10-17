package bridge

import (
	"fmt"

	"github.com/TheThingsIndustries/mystique/pkg/log"
	"google.golang.org/grpc/grpclog"
)

func SetGRPCLogger(log log.Interface) {
	grpclog.SetLoggerV2(gRPCLogger{log})
}

type gRPCLogger struct {
	log.Interface
}

func (l gRPCLogger) Info(args ...interface{}) {
	l.Interface.Debug(fmt.Sprint(args...))
}
func (l gRPCLogger) Infoln(args ...interface{}) {
	l.Interface.Debug(fmt.Sprint(args...))
}
func (l gRPCLogger) Infof(format string, args ...interface{}) {
	l.Interface.Debugf(format, args...)
}
func (l gRPCLogger) Warning(args ...interface{}) {
	l.Interface.Warn(fmt.Sprint(args...))
}
func (l gRPCLogger) Warningln(args ...interface{}) {
	l.Interface.Warn(fmt.Sprint(args...))
}
func (l gRPCLogger) Warningf(format string, args ...interface{}) {
	l.Interface.Warnf(format, args...)
}
func (l gRPCLogger) Error(args ...interface{}) {
	l.Interface.Error(fmt.Sprint(args...))
}
func (l gRPCLogger) Errorln(args ...interface{}) {
	l.Interface.Error(fmt.Sprint(args...))
}
func (l gRPCLogger) Fatal(args ...interface{}) {
	l.Interface.Fatal(fmt.Sprint(args...))
}
func (l gRPCLogger) Fatalln(args ...interface{}) {
	l.Interface.Fatal(fmt.Sprint(args...))
}

func (l gRPCLogger) V(_ int) bool { return true }
