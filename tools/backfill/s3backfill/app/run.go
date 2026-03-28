package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func Run() {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:9000"
	}
	accessKeyID := os.Getenv("MINIO_ROOT_USER")
	if accessKeyID == "" {
		accessKeyID = "admin"
	}
	secretAccessKey := os.Getenv("MINIO_ROOT_PASSWORD")
	if secretAccessKey == "" {
		secretAccessKey = "MudroAdmin2026"
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "media"
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: "us-east-1"})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("Bucket '%s' already exists\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created bucket '%s'\n", bucketName)
		policy := fmt.Sprintf(`{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::%s/*"],"Sid": ""}]}`, bucketName)
		err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			log.Printf("Failed to set public policy for bucket: %v", err)
		} else {
			log.Println("Bucket policy set to public read")
		}
	}

	mediaRoot := strings.TrimSpace(os.Getenv("MEDIA_ROOT"))
	if mediaRoot == "" {
		mediaRoot = filepath.Join(config.RepoRoot(), "data", "nu")
	}

	log.Printf("Scanning directory: %s", mediaRoot)
	var count, uploaded int
	err = filepath.Walk(mediaRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		
		relPath, err := filepath.Rel(mediaRoot, path)
		if err != nil {
			return err
		}
		objectName := strings.ReplaceAll(relPath, "\\", "/")
		
		// Upload the file
		_, putErr := minioClient.FPutObject(ctx, bucketName, objectName, path, minio.PutObjectOptions{})
		if putErr != nil {
			log.Printf("Failed to upload %s: %v", objectName, putErr)
		} else {
			uploaded++
			if uploaded%100 == 0 {
				log.Printf("Uploaded %d files...", uploaded)
			}
		}
		count++
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the path %v: %v\n", mediaRoot, err)
	}
	log.Printf("Finished. Discovered %d files, uploaded %d successfully.", count, uploaded)
}
