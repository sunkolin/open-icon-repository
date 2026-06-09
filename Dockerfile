# 构建阶段
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod ./
RUN go mod download

# 复制源代码
COPY main.go ./
COPY index.html ./
COPY icon ./icon

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o open-icon-repository .

# 运行阶段
FROM alpine:latest

# 安装必要的工具
RUN apk add --no-cache wget

WORKDIR /app

# 从构建阶段复制二进制文件和静态资源
COPY --from=builder /app/open-icon-repository .
COPY --from=builder /app/index.html .
COPY --from=builder /app/icon ./icon

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./open-icon-repository"]
