FROM ubuntu:22.04
ENV GOOS=linux
ENV CGO_ENABLED=1
ARG S3_REGION
ARG S3_BUCKET
ARG S3_ACCESS_KEY
ARG S3_SECRET_KEY
WORKDIR /lambda
RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo Asia/Shanghai > /etc/timezone && \
    apt update && apt upgrade -y && apt install -y git gcc g++ ca-certificates golang
RUN wget https://github.com/kahing/goofys/releases/latest/download/goofys && \
    chmod +x goofys && mv goofys /usr/local/bin && \
    echo "[default]\naws_access_key_id = ${S3_ACCESS_KEY}\naws_secret_access_key = ${S3_SECRET_KEY}\n" > /root/.aws/credentials && \
    goofys --region ${S3_REGION} ${S3_BUCKET} /lambda/warehouse

COPY . .
ENV GO_DYNAMIC_WAREHOUSE=/lambda/warehouse
RUN go mod download && go build -o bootstrap ./httpserver

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.5.0 /lambda-adapter /opt/extensions/lambda-adapter

ENV PORT=8000 GIN_MODE=release
EXPOSE 8000

CMD ["/lambda/bootstrap"]
