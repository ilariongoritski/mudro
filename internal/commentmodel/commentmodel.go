package commentmodel

import (
	"context"
	"errors"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func SyncCommentReactions(ctx context.Context, q Querier, commentID int64, reactions map[string]int) error {
	if _, err := q.Exec(ctx, `delete from comment_reactions where comment_id = $1`, commentID); err != nil {
		if IsUndefinedTableErr(err) {
			return nil
		}
		return err
	}

	if len(reactions) == 0 {
		return nil
	}

	keys := make([]string, 0, len(reactions))
	for emoji, count := range reactions {
		if count > 0 {
			keys = append(keys, emoji)
		}
	}
	sort.Strings(keys)

	for _, emoji := range keys {
		count := reactions[emoji]
		if _, err := q.Exec(ctx, `
insert into comment_reactions (comment_id, emoji, count, updated_at)
values ($1, $2, $3, now())
on conflict (comment_id, emoji) do update set
  count = excluded.count,
  updated_at = now()
`, commentID, emoji, count); err != nil {
			return err
		}
	}
	return nil
}

func LoadCommentReactions(ctx context.Context, q Querier, commentIDs []int64) (map[int64]map[string]int, error) {
	out := make(map[int64]map[string]int, len(commentIDs))
	if len(commentIDs) == 0 {
		return out, nil
	}

	rows, err := q.Query(ctx, `
select comment_id, emoji, count
from comment_reactions
where comment_id = any($1)
order by comment_id asc, emoji asc
`, commentIDs)
	if err != nil {
		if IsUndefinedTableErr(err) {
			return out, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			commentID int64
			emoji     string
			count     int
		)
		if err := rows.Scan(&commentID, &emoji, &count); err != nil {
			return nil, err
		}
		if count <= 0 {
			continue
		}
		if out[commentID] == nil {
			out[commentID] = map[string]int{}
		}
		out[commentID][emoji] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func IsUndefinedTableErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}
