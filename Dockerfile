# buildkit跨架构编译缓慢，若未使用CGO，可统一使用本机架构：BUILDPLATFORM，进行交叉编译
FROM --platform=$BUILDPLATFORM docker.io/wbuntu/golang:1.21 AS builder
ARG TARGETARCH
ENV GOARCH=$TARGETARCH
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GOPROXY="https://goproxy.cn,direct"
COPY . /free-ask-bot
WORKDIR /free-ask-bot
RUN make build

# 基础镜像
FROM --platform=$TARGETPLATFORM docker.io/wbuntu/alpine:3.18
# 固定时区为东八区
ENV TZ=Asia/Shanghai
# 拷贝二进制文件
COPY --from=builder /free-ask-bot/build/free-ask-bot /usr/bin/free-ask-bot
CMD ["/usr/bin/free-ask-bot","-c","/etc/free-ask-bot/config.toml"]