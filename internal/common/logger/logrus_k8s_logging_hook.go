package logger

import (
	"encoding/json"
	"fmt"
	sysLog "log"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	rolling_logger "gopkg.in/natefinch/lumberjack.v2"

	"github.com/amazingchow/wechat-payment-callback-service/internal/service/common"
)

var _LogrusAlarmLevels = []logrus.Level{
	logrus.InfoLevel,
	logrus.WarnLevel,
	logrus.ErrorLevel,
	logrus.FatalLevel,
	logrus.PanicLevel,
}

type LogrusKubernatesLog struct {
	event string
	msg   string
	kvs   map[string]interface{}
}

type LogrusKubernatesLoggingHook struct {
	app   string
	group string

	filePath       string
	maxFileSize    int // MB
	maxFileBackups int

	l         *rolling_logger.Logger
	logBuffer chan *LogrusKubernatesLog
}

var _Hook *LogrusKubernatesLoggingHook

func WithApp(x string) func(*LogrusKubernatesLoggingHook) {
	return func(hook *LogrusKubernatesLoggingHook) {
		if len(x) > 0 {
			hook.app = x
		}
	}
}

func WithGroup(x string) func(*LogrusKubernatesLoggingHook) {
	return func(hook *LogrusKubernatesLoggingHook) {
		if len(x) > 0 {
			hook.group = x
		}
	}
}

func WithFilePath(x string) func(*LogrusKubernatesLoggingHook) {
	return func(hook *LogrusKubernatesLoggingHook) {
		if len(x) > 0 {
			hook.filePath = x
		}
	}
}

func WithMaxFileSize(x int) func(*LogrusKubernatesLoggingHook) {
	return func(hook *LogrusKubernatesLoggingHook) {
		if x > 0 {
			hook.maxFileSize = x
		}
	}
}

func WithMaxFileBackups(x int) func(*LogrusKubernatesLoggingHook) {
	return func(hook *LogrusKubernatesLoggingHook) {
		if x > 0 {
			hook.maxFileBackups = x
		}
	}
}

func SetupLogrusKubernatesLoggingHook(opts ...func(*LogrusKubernatesLoggingHook)) {
	_Hook = &LogrusKubernatesLoggingHook{}
	for _, opt := range opts {
		opt(_Hook)
	}
	if len(_Hook.app) == 0 {
		_Hook.app = "my-app"
	}
	if len(_Hook.group) == 0 {
		_Hook.group = "my-group"
	}
	if len(_Hook.filePath) == 0 {
		_Hook.filePath = "/app/logs"
	}
	if _Hook.maxFileSize == 0 {
		_Hook.maxFileSize = 64
	}
	if _Hook.maxFileBackups == 0 {
		_Hook.maxFileBackups = 4
	}
	_Hook.l = &rolling_logger.Logger{
		Filename:   fmt.Sprintf("%s/%s-%s.log", _Hook.filePath, _Hook.app, _Hook.group),
		MaxSize:    _Hook.maxFileSize,
		MaxBackups: _Hook.maxFileBackups,
	}
	_Hook.logBuffer = make(chan *LogrusKubernatesLog, 1024)

	go _Hook._RunAsyncCommit()
}

func GetLogrusKubernatesLoggingHook() *LogrusKubernatesLoggingHook {
	return _Hook
}

func CloseLogrusKubernatesLoggingHook() {
	close(_Hook.logBuffer)
	_Hook.l.Close()
}

func (hook *LogrusKubernatesLoggingHook) genRow(event string, msg string, kvs map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(kvs)+5)
	for k, v := range kvs {
		m[k] = v
	}
	// These fields must not be modified by kvs
	m["app"] = hook.app
	m["group"] = hook.group
	m["event"] = event
	m["msg"] = msg
	m["ts"] = Now()
	return m
}

func (hook *LogrusKubernatesLoggingHook) dumpRow(row map[string]interface{}) error {
	line, err := json.Marshal(row)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal current log.")
	}
	line = append(line, []byte("\n")...)
	_, err = hook.l.Write(line)
	if err != nil {
		return errors.Wrapf(err, "Failed to persist current log to file:%s.", hook.filePath)
	}
	return nil
}

func (hook *LogrusKubernatesLoggingHook) _RunAsyncCommit() {
	var err error

	for log := range hook.logBuffer {
		row := hook.genRow(log.event, log.msg, log.kvs)
		if err = hook.dumpRow(row); err != nil {
			sysLog.Println(err)
		}
	}
}

func (hook *LogrusKubernatesLoggingHook) Fire(entry *logrus.Entry) error {
	var event string
	_evt, ok := entry.Data[common.LoggerKeyEvent]
	if ok {
		event = _evt.(string)
	} else {
		event = "no-event"
	}

	var msg string
	err, ok := entry.Data[logrus.ErrorKey]
	if ok {
		msg = fmt.Sprintf("%s, err: %v", entry.Message, err)
	} else {
		msg = entry.Message
	}

	kvs := make(map[string]interface{}, 2)
	_traceId, ok := entry.Data[common.LoggerKeyTraceId]
	if ok {
		kvs["trace-id"] = _traceId.(string)
	}
	_spanId, ok := entry.Data[common.LoggerKeySpanId]
	if ok {
		kvs["span-id"] = _spanId.(string)
	}

	hook.logBuffer <- &LogrusKubernatesLog{
		event: event,
		msg:   msg,
		kvs:   kvs,
	}
	return nil
}

func (hook *LogrusKubernatesLoggingHook) Levels() []logrus.Level {
	return _LogrusAlarmLevels
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}

func Now() string {
	return FormatTime(time.Now())
}
