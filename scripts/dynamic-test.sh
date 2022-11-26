export dynamic_version="v1.0.2"
docker build -t lambda-dynamic-cli:${dynamic_version} -f ./docker/dynamic-cli.Dockerfile --build-arg VERSION=${dynamic_version} .
docker run -it --rm -v `pwd`/warehouse:/opt/go-dynamic-warehouse lambda-dynamic-cli:${dynamic_version} build github.com/aura-studio/testdynamic1@test
docker run -it --rm -v `pwd`/warehouse:/opt/go-dynamic-warehouse lambda-dynamic-cli:${dynamic_version} build github.com/aura-studio/testdynamic2@test

docker build -t lambda-http-server -f ./docker/http-server.Dockerfile .
docker stop lambda-http-server || true
docker run -it --rm --name lambda-http-server -p 8080:8000 -v `pwd`/warehouse:/opt/go-dynamic-warehouse lambda-http-server