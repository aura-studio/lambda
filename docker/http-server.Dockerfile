FROM ubuntu:22.04 as builder
ENV GOOS=linux
ENV CGO_ENABLED=1
WORKDIR /lambda
RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo Asia/Shanghai > /etc/timezone && \
    apt update && apt upgrade -y && apt install -y git gcc g++ ca-certificates golang
COPY . .
RUN go mod download && go build -o bootstrap ./httpserver


FROM ubuntu:22.04
COPY --from=builder /lambda/bootstrap /lambda/bootstrap
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.5.0 /lambda-adapter /opt/extensions/lambda-adapter
ENV PORT=8000 GIN_MODE=release GO_DYNAMIC_WAREHOUSE=/tmp/warehouse

EXPOSE 8000

CMD ["/lambda/bootstrap"]
