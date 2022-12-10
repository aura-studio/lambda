docker volume create scylla-data
docker run -d --name scylla --hostname scylla --restart=always -p 8001:8001 -v scylla-data:/var/lib/scylla scylladb/scylla --alternator-port 8001
