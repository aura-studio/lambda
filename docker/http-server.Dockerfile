FROM ubuntu:22.04
ENV GOOS=linux
ENV CGO_ENABLED=1
WORKDIR /lambda
RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo Asia/Shanghai > /etc/timezone && \
    apt update && apt upgrade -y && apt install -y git gcc g++ ca-certificates golang

COPY . .
ENV GO_DYNAMIC_WAREHOUSE=/lambda/warehouse
RUN go mod download && go build -o bootstrap ./httpserver

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.5.0 /lambda-adapter /opt/extensions/lambda-adapter

ENV PORT=8000 GIN_MODE=release
EXPOSE 8000

CMD ["/lambda/bootstrap"]
