FROM golang:1.20-alpine AS builder

WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod ./
COPY go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o bt_search .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 复制可执行文件和必要的资源
COPY --from=builder /app/bt_search .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/config ./config

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./bt_search"]
