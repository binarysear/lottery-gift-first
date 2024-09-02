#!/bin/sh

# 日志函数
log() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 等待 MySQL 准备好
log "Waiting for MySQL to be ready..."
until mysqladmin ping -h mysql -u root -proot --silent; do
  log "MySQL not ready yet, retrying in 5 seconds..."
  sleep 5
done
log "MySQL is ready."

# 等待 RocketMQ Broker 准备好
log "Waiting for RocketMQ Broker to be ready..."
until nc -z rmqbroker 10911; do
  log "RocketMQ Broker not ready yet, retrying in 5 seconds..."
  sleep 5
done
log "RocketMQ Broker is ready."

sleep 20

# 一旦所有服务都准备好，启动 Go 程序
log "All services are ready, starting Go application..."
exec ./app
