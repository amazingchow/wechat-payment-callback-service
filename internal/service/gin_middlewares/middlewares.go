package middlewares

import (
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/logger"
	"github.com/amazingchow/wechat-payment-callback-service/internal/service/common"
)

func RecoverPanicAndReportLatencyMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		tid := uuid.New().String()
		sid := uuid.New().String()
		v := ctx.GetHeader(common.ReqHeaderKeyRequestId)
		if len(v) > 0 {
			tid = v
		}
		ctx.Set(string(common.ContextKeyTraceId), tid)
		ctx.Set(string(common.ContextKeySpanId), sid)

		st := time.Now()
		defer func() {
			ed := time.Now()
			logger.GetGlobalLogger().
				WithField(common.LoggerKeyTraceId, tid).
				WithField(common.LoggerKeySpanId, sid).
				WithField("latency", ed.Sub(st).Milliseconds()).
				Debug("request latency")

			if e := recover(); e != nil {
				logger.GetGlobalLogger().
					WithField(common.LoggerKeyTraceId, tid).
					WithField(common.LoggerKeySpanId, sid).
					WithField("stack", string(debug.Stack())).
					WithError(e.(error)).
					Error("Recover from internal server panic.")
				ctx.AbortWithStatus(500)
			}
		}()

		ctx.Next()
	}
}
