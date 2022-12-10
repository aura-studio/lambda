docker volume create scylla-data
docker run --name scylla --restart=always -d -p 8000:8000 -v scylla-data:/var/lib/scylla scylladb/scylla-nightly:latest --alternator-port=8000 --alternator-write-isolation=always
