package interceptor

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/logger"
	"github.com/amazingchow/wechat-payment-callback-service/internal/service/common"
)

func RecoverPanicAndReportLatencyUnaryInterceptor(
	ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, rpcHandler grpc.UnaryHandler) (
	resp interface{}, err error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = status.Error(codes.InvalidArgument, "Request header must be provided.")
		return
	}

	tid := uuid.New().String()
	sid := uuid.New().String()
	if v, ok := md[common.ReqHeaderKeyRequestId]; ok {
		tid = v[0]
	}
	ctx = context.WithValue(ctx, common.ContextKeyTraceId, tid)
	ctx = context.WithValue(ctx, common.ContextKeySpanId, sid)

	st := time.Now()
	defer func() {
		if e := recover(); e != nil {
			logger.GetGlobalLogger().
				WithField(common.LoggerKeyTraceId, tid).
				WithField(common.LoggerKeySpanId, sid).
				WithField("stack", string(debug.Stack())).
				WithError(e.(error)).
				Error("Recover from internal server panic.")
			err = status.Error(codes.Internal, "Recover from internal server panic.")
		}

		ed := time.Now()
		logger.GetGlobalLogger().
			WithField(common.LoggerKeyTraceId, tid).
			WithField(common.LoggerKeySpanId, sid).
			WithField("latency", ed.Sub(st).Milliseconds()).
			Debug("request latency")
	}()

	return rpcHandler(ctx, req)
}
