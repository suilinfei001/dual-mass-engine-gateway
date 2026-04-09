# AI Analyzer Dockerfile
# 使用预编译的本地二进制文件（静态编译）
FROM acr.aishu.cn/dual-mass-engine-gateway/alpine:3.19

WORKDIR /app

# 安装 CA 证书和时区数据

# 复制本地编译好的二进制文件（由 deploy.sh 编译）
COPY output/ai-analyzer /app/ai-analyzer

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 4005

# 启动服务
CMD ["/app/ai-analyzer"]
