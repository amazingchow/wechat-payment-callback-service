package logger

import (
	"github.com/evalphobia/logrus_sentry"
	"github.com/sirupsen/logrus"

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/config"
)

var _Logger *logrus.Entry

func SetGlobalLogger(conf *config.Config) {
	var _logger *logrus.Logger
	var _level logrus.Level
	var err error

	_logger = logrus.New()
	// Set log level.
	if len(conf.LogLevel) > 0 {
		_level, err = logrus.ParseLevel(conf.LogLevel)
		if err != nil {
			_level = logrus.DebugLevel
		}
	} else {
		_level = logrus.DebugLevel
	}
	_logger.SetLevel(_level)
	// Set log formatter.
	_logger.SetFormatter(&logrus.TextFormatter{})
	// Set hook for sentry.
	if len(conf.LogSentryDSN) > 0 {
		hook, err := logrus_sentry.NewSentryHook(conf.LogSentryDSN, []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to set sentry hook for logrus logger.")
		} else {
			_logger.Hooks.Add(hook)
		}
	}
	if conf.LogPrinter == "disk" {
		// Set hook for disk.
		opts := make([]func(*LogrusKubernatesLoggingHook), 0, 5)
		opts = append(opts, WithApp(conf.ServiceName))
		opts = append(opts, WithGroup(conf.ServiceGroupName))
		opts = append(opts, WithFilePath(conf.LogPrinterFilePath))
		SetupLogrusKubernatesLoggingHook(opts...)
		hook := GetLogrusKubernatesLoggingHook()
		_logger.Hooks.Add(hook)
	}
	// Set log fields.
	_Logger = _logger.WithField("service_name", conf.ServiceName)
}

func GetGlobalLogger() *logrus.Entry {
	return _Logger
}
