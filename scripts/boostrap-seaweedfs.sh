docker volume create seaweedfs-data
docker volume create seaweedfs-buckets
docker run -d --name seaweedfs --hostname seaweedfs --restart=always -p 8333:8333 -v seaweedfs-buckets:/buckets -v seawedfs-data:/data chrislusf/seaweedfs server -s3