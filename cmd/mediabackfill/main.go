package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	mediadb "github.com/goritskimihail/mudro/internal/media"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := flag.String("dsn", config.DSN(), "postgres dsn")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	postCount, err := backfillPosts(ctx, pool)
	if err != nil {
		log.Fatalf("backfill posts: %v", err)
	}
	commentCount, err := backfillComments(ctx, pool)
	if err != nil {
		log.Fatalf("backfill comments: %v", err)
	}

	log.Printf("DONE: post rows synced=%d comment rows synced=%d", postCount, commentCount)
}

func backfillPosts(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	rows, err := pool.Query(ctx, `
select id, source, media
from posts
where media is not null
order by id asc
`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var (
			postID int64
			source string
			raw    []byte
		)
		if err := rows.Scan(&postID, &source, &raw); err != nil {
			return count, err
		}
		if err := mediadb.SyncPostLinks(ctx, pool, postID, source, json.RawMessage(raw)); err != nil {
			return count, err
		}
		count++
	}
	return count, rows.Err()
}

func backfillComments(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	rows, err := pool.Query(ctx, `
select id, source, media
from post_comments
where media is not null
order by id asc
`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var (
			commentID int64
			source    string
			raw       []byte
		)
		if err := rows.Scan(&commentID, &source, &raw); err != nil {
			return count, err
		}
		if err := mediadb.SyncCommentLinks(ctx, pool, commentID, source, json.RawMessage(raw)); err != nil {
			return count, err
		}
		count++
	}
	return count, rows.Err()
}
