FROM golang:1.21.10 as builder
WORKDIR /build
ADD . /build/
RUN make build

FROM alpine:3.18
WORKDIR /app
RUN apk add libc6-compat
COPY --from=builder /build/wechat-payment-callback-service /app
RUN mkdir -p /app/config /app/logs /app/persistent /app/locks /app/shares
EXPOSE 16887
EXPOSE 16888
CMD [ \
    "/app/wechat-payment-callback-service", \
    "-conf", \
    "/config/wechat-payment-callback-service.json" \
]
