package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/config"
	"github.com/amazingchow/wechat-payment-callback-service/internal/common/logger"
	ext_redis "github.com/amazingchow/wechat-payment-callback-service/internal/extensions/ext_redis"
	"github.com/amazingchow/wechat-payment-callback-service/internal/metrics"
	"github.com/amazingchow/wechat-payment-callback-service/internal/service"
)

var (
	_ConfigFile = flag.String("conf", "./etc/wechat-payment-callback-service-dev.json", "config file path")
)

func main() {
	// Seed random number generator, is deprecated since Go 1.20.
	// rand.Seed(time.Now().UnixNano())
	logrus.SetLevel(logrus.DebugLevel)

	flag.Parse()
	config.LoadConfigFileOrPanic(*_ConfigFile)
	defer SetupTeardown()()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopCh := make(chan struct{})
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	wg.Add(1)
	go setupGrpcService(ctx, wg, stopCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	close(stopCh)
}

func SetupTeardown() func() {
	logrus.Debug("Run service-initialization.")
	SetupRuntimeEnvironment(config.GetConfig())
	return func() {
		logrus.Debug("Run service-cleanup.")
		ClearRuntimeEnvironment(config.GetConfig())
	}
}

func SetupRuntimeEnvironment(conf *config.Config) {
	logger.SetGlobalLogger(conf)
	if len(conf.ServiceMetricsEndpoint) > 0 {
		go func() {
			metrics.Register()
			http.Handle("/metrics", promhttp.Handler())
			logger.GetGlobalLogger().Error(http.ListenAndServe(conf.ServiceMetricsEndpoint, nil))
		}()
	}
	// Add more service initialization here.
	ext_redis.InitConnPool(&conf.ServiceInternalConfig.Cache)
	service.SetupWechatPaymentCallbackServiceImpl()
}

func ClearRuntimeEnvironment(_ *config.Config) {
	service.CloseWechatPaymentCallbackServiceImpl()
	// Add more service cleanup here.
	ext_redis.CloseConnPool()
}
