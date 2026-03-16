package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	commentdb "github.com/goritskimihail/mudro/internal/commentmodel"
	"github.com/goritskimihail/mudro/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := flag.String("dsn", config.DSN(), "postgres dsn")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	parentLinks, err := backfillParentCommentIDs(ctx, pool)
	if err != nil {
		log.Fatalf("backfill parent_comment_id: %v", err)
	}
	reactionRows, err := backfillCommentReactions(ctx, pool)
	if err != nil {
		log.Fatalf("backfill comment_reactions: %v", err)
	}

	log.Printf("DONE: parent_links=%d reaction_rows_synced=%d", parentLinks, reactionRows)
}

func backfillParentCommentIDs(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	tag, err := pool.Exec(ctx, `
update post_comments child
set parent_comment_id = parent.id,
    updated_at = now()
from post_comments parent
where child.source_parent_comment_id is not null
  and child.source = parent.source
  and child.post_id = parent.post_id
  and parent.source_comment_id = child.source_parent_comment_id
  and child.parent_comment_id is distinct from parent.id
`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func backfillCommentReactions(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	rows, err := pool.Query(ctx, `
select id, reactions
from post_comments
where reactions is not null
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
			raw       []byte
			reactions map[string]int
		)
		if err := rows.Scan(&commentID, &raw); err != nil {
			return count, err
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &reactions); err != nil {
				return count, err
			}
		}
		if err := commentdb.SyncCommentReactions(ctx, pool, commentID, reactions); err != nil {
			return count, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return count, err
	}
	return count, nil
}
