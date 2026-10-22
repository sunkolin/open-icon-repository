# 构建阶段
FROM ubuntu:26.04 AS builder

RUN apt-get update && \
    apt-get install -y --no-install-recommends golang-go ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY index.html ./
COPY icon ./icon
COPY baitutubiao.png ./

RUN CGO_ENABLED=0 GOOS=linux go build -o open-icon-repository .

# 运行阶段
FROM ubuntu:26.04

ENV TZ=Asia/Shanghai

RUN apt-get update && \
    apt-get install -y --no-install-recommends wget ca-certificates curl telnet tzdata vim && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone && \
    echo "alias ll='ls -al'" >> /root/.bashrc && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/open-icon-repository .
COPY --from=builder /app/index.html .
COPY --from=builder /app/icon ./icon
COPY --from=builder /app/baitutubiao.png ./

EXPOSE 6024

CMD ["./open-icon-repository"]