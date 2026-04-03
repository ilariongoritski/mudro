package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/goritskimihail/mudro/internal/media"
)

func main() {
	ctx := context.Background()

	// 1. MinIO Init
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:9000"
	}
	accessKey := os.Getenv("MINIO_ROOT_USER")
	if accessKey == "" {
		accessKey = "admin"
	}
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	if secretKey == "" {
		secretKey = "MudroAdmin2026"
	}
	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "mudro-media"
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("MinIO client: %v", err)
	}

	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Check bucket: %v", err)
	}
	if !exists {
		log.Printf("Creating bucket %s...", bucketName)
		if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("Make bucket: %v", err)
		}
		policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetBucketLocation","s3:ListBucket"],"Resource":["arn:aws:s3:::%s"]},{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucketName, bucketName)
		if err := minioClient.SetBucketPolicy(ctx, bucketName, policy); err != nil {
			log.Printf("Warning: failed to set policy: %v", err)
		}
	}

	// 2. Postgres Connection
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@127.0.0.1:5433/gallery?sslmode=disable"
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Postgres connection: %v", err)
	}
	defer pool.Close()

	// 3. Backfill Posts
	log.Println("Backfilling posts media...")
	rows, err := pool.Query(ctx, "select id, source, media from posts where media is not null")
	if err != nil {
		log.Fatalf("Query posts: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var source string
		var raw json.RawMessage
		if err := rows.Scan(&id, &source, &raw); err != nil {
			log.Printf("Scan post %d: %v", id, err)
			continue
		}
		if source == "" {
			source = "native"
		}
		if err := media.SyncPostLinks(ctx, pool, id, source, raw); err != nil {
			log.Printf("Sync post %d: %v", id, err)
		}
	}

	// 4. Backfill Comments
	log.Println("Backfilling comments media...")
	crows, err := pool.Query(ctx, "select id, media from post_comments where media is not null")
	if err != nil {
		log.Printf("Query comments: %v", err)
	} else {
		defer crows.Close()
		for crows.Next() {
			var id int64
			var raw json.RawMessage
			if err := crows.Scan(&id, &raw); err != nil {
				log.Printf("Scan comment %d: %v", id, err)
				continue
			}
			if err := media.SyncCommentLinks(ctx, pool, id, "native", raw); err != nil {
				log.Printf("Sync comment %d: %v", id, err)
			}
		}
	}

	log.Println("Backfill complete.")
}
