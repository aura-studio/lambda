docker build -t lambda-http-server -f ./docker/http-server-ubuntu.Dockerfile .
docker run -it --rm --name lambda-http-server -p 8080:8000 -e GO_DYNAMIC_WAREHOUSE=/opt/go-dynamic-warehouse lambda-http-server