docker volume create scylla-data
docker run -d --name scylla --hostname scylla --restart=always -p 9042:9042 -v scylla-data:/var/lib/scylla scylladb/scylla