FROM ubuntu:22.04
ENV GOOS=linux
ENV CGO_ENABLED=1
RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo Asia/Shanghai > /etc/timezone && \
    apt update && apt upgrade -y && apt install -y git gcc g++ ca-certificates golang