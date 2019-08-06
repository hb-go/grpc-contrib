package zap

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/grpclog"
)

var (
	SystemField = zap.String("system", "grpc")
)

// ReplaceGrpcLogger sets the given zap.Logger as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
func ReplaceGrpcLogger(logger *zap.Logger) {
	zgl := &zapLogger{logger.With(SystemField, zap.Bool("grpc_log", true))}
	grpclog.SetLoggerV2(zgl)
}

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Info(v ...interface{}) {
	l.logger.Info(fmt.Sprint(v...))
}

func (l *zapLogger) Infoln(v ...interface{}) {
	l.logger.Info(fmt.Sprintln(v...))
}

func (l *zapLogger) Infof(format string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, v...))
}

func (l *zapLogger) Warning(v ...interface{}) {
	l.logger.Warn(fmt.Sprint(v...))
}

func (l *zapLogger) Warningln(v ...interface{}) {
	l.logger.Warn(fmt.Sprintln(v...))
}

func (l *zapLogger) Warningf(format string, v ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, v...))
}

func (l *zapLogger) Error(v ...interface{}) {
	l.logger.Error(fmt.Sprint(v...))
}

func (l *zapLogger) Errorln(v ...interface{}) {
	l.logger.Error(fmt.Sprintln(v...))
}

func (l *zapLogger) Errorf(format string, v ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, v...))
}

func (l *zapLogger) Fatal(v ...interface{}) {
	l.logger.Fatal(fmt.Sprint(v...))
}

func (l *zapLogger) Fatalln(v ...interface{}) {
	l.logger.Fatal(fmt.Sprintln(v...))
}

func (l *zapLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(format, v...))
}

func (l *zapLogger) V(lvl int) bool {
	return true
}
