# Webhook Gateway Dockerfile
# 使用预编译的本地二进制文件（静态编译）
FROM acr.aishu.cn/dual-mass-engine-gateway/alpine:3.19

WORKDIR /app

# 复制本地编译好的二进制文件（由 deploy.sh 编译）
COPY output/webhook-gateway /app/webhook-gateway

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 4001

# 启动服务
CMD ["/app/webhook-gateway"]
