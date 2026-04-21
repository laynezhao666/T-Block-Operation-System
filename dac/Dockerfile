# 第一次构建: 二进制可执行文件
FROM golang:1.23-bookworm AS builder
WORKDIR /data/workspace/
## 复制代码至相关文件
COPY . /data/workspace/
## 下载依赖
RUN go mod tidy
## 编译
RUN mkdir cmd && go build -o /data/workspace/cmd/tdac

# 第二次构建: 部署二进制可执行文件并执行
FROM debian:bookworm-slim
WORKDIR /root
## 从第一次构建的命令中获取所需文件
COPY --from=builder /data/workspace/cmd/tdac /root

RUN apt-get update && apt-get install -y --no-install-recommends \
    net-tools tcpdump \
    && rm -rf /var/lib/apt/lists/*

## 设置启动命令
CMD ["./tdac"]

