docker volume create scylla-data
docker run --name scylla --restart=always -d -p 18000:18000 -v scylla-data:/var/lib/scylla scylladb/scylla-nightly:latest --alternator-port=18000 --alternator-write-isolation=always
