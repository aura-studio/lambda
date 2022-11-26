FROM ubuntu:22.04
ENV GOOS=linux
ENV CGO_ENABLED=1
RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo Asia/Shanghai > /etc/timezone && \
    apt update && apt upgrade -y && apt install -y git gcc g++ ca-certificates golang

ARG VERSION
RUN	go install github.com/aura-studio/dynamic/dynamic-cli@${VERSION}

ENTRYPOINT ["/root/go/bin/dynamic-cli"]
