# Auth Service Dockerfile
FROM ubuntu:22.04

WORKDIR /app

# 复制编译好的二进制文件
COPY auth-service /app/auth-service

RUN chmod +x /app/auth-service

# 设置权限
RUN chmod +x /app/auth-service

# 暴露端口
EXPOSE 4007

# 设置环境变量默认值
ENV ADMIN_USER=admin
ENV ADMIN_PASS=admin123
ENV PORT=4007

# 运行服务
CMD ["/app/auth-service"]
