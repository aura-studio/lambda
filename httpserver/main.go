package main

import (
	"fmt"
	"os"

	"github.com/aura-studio/dynamic"
	"github.com/aura-studio/lambda/engine"
)

func init() {
	fmt.Println(dynamic.S3Config{
		Region:    os.Getenv("S3_REGION"),
		Bucket:    os.Getenv("S3_BUCKET"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	})

	// These environment variables must be set:
	// S3_REGION/S3_BUCKET/S3_ACCESS_KEY/S3_SECRET_KEY
	dynamic.WithRemote(dynamic.NewS3Remote(&dynamic.S3Config{
		Region:    os.Getenv("S3_REGION"),
		Bucket:    os.Getenv("S3_BUCKET"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	}))
}

func main() {
	engine.ServeHTTP()
}
