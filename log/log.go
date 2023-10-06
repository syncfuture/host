package log

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Lukiya/logs"
	"github.com/kataras/golog"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/sgrpc"
)

var (
	_logServiceClient logs.LogEntryServiceClient
	_sbPool           = &sync.Pool{
		New: func() any {
			return new(strings.Builder)
		},
	}
)

func UseGrpcLogs(consulAddr string, consulToken map[string]string, clientID string) {
	logServiceConn, err := sgrpc.DialConsul(consulAddr, "logs", consulToken)
	u.LogFatal(err)
	_logServiceClient = logs.NewLogEntryServiceClient(logServiceConn)

	golog.Handle(func(value *golog.Log) (handled bool) {
		grpcLogSink(value, clientID)
		return false
	})
}

func grpcLogSink(value *golog.Log, clientID string) {
	sb := _sbPool.Get().(*strings.Builder)
	sb.Reset()
	defer _sbPool.Put(sb)

	if len(value.Stacktrace) > 0 {
		for _, x := range value.Stacktrace {
			sb.WriteString(x.String())
		}
	}

	_logServiceClient.WriteLogEntry(context.Background(), &logs.WriteLogCommand{
		ClientID: clientID,
		LogEntry: &logs.LogEntry{
			Level:        convertLevel(value.Level),
			Message:      value.Message,
			StackTrace:   sb.String(),
			CreatedOnUtc: time.Now().UTC().UnixMilli(),
		},
	})
}

func convertLevel(level golog.Level) logs.LogLevel {
	switch level {
	case golog.DebugLevel:
		return logs.LogLevel_Debug
	case golog.InfoLevel:
		return logs.LogLevel_Infomation
	case golog.WarnLevel:
		return logs.LogLevel_Warning
	case golog.ErrorLevel:
		return logs.LogLevel_Error
	case golog.FatalLevel:
		return logs.LogLevel_Fatal
	default:
		return logs.LogLevel_Verbose
	}
}
