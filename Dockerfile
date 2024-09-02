# 第一阶段: 构建
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
ENV GOPROXY https://goproxy.cn,direct
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# 第二阶段: 生产环境镜像
FROM alpine:3.16

RUN apk update --no-cache && apk add --no-cache ca-certificates && apk add --no-cache mariadb-client && apk add --no-cache redis

WORKDIR /app
COPY --from=builder /app/app .
COPY config /app/config
ADD views /app/views
RUN mkdir -p /app/log
# 添加健康检查脚本
COPY wait-for-services.sh /app/wait-for-services.sh
RUN chmod +x /app/wait-for-services.sh

EXPOSE 5679

# 启动健康检查脚本
CMD ["/app/wait-for-services.sh"]
