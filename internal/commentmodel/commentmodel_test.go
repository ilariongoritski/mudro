package commentmodel

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockQuerier implements Querier for testing
type mockQuerier struct {
	execFn  func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	queryFn func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func (m *mockQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execFn != nil {
		return m.execFn(ctx, sql, args...)
	}
	return pgconn.NewCommandTag(""), nil
}

func (m *mockQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, sql, args...)
	}
	return nil, nil
}

func TestSyncCommentReactions_NoReactions(t *testing.T) {
	mq := &mockQuerier{}
	err := SyncCommentReactions(context.Background(), mq, 1, map[string]int{})
	if err != nil {
		t.Errorf("SyncCommentReactions with empty map: %v", err)
	}
}

func TestSyncCommentReactions_Insert(t *testing.T) {
	var execCount int
	mq := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			execCount++
			if execCount == 1 {
				// DELETE
				if sql != "delete from comment_reactions where comment_id = $1" {
					t.Errorf("unexpected SQL: %s", sql)
				}
				if args[0] != int64(1) {
					t.Errorf("unexpected args: %v", args)
				}
			} else {
				// INSERT
				if args[0] != int64(1) || args[1] != "👍" || args[2] != 5 {
					t.Errorf("unexpected INSERT args: %v", args)
				}
			}
			return pgconn.NewCommandTag("INSERT 0 1"), nil
		},
	}

	err := SyncCommentReactions(context.Background(), mq, 1, map[string]int{"👍": 5})
	if err != nil {
		t.Errorf("SyncCommentReactions: %v", err)
	}
	if execCount != 2 {
		t.Errorf("expected 2 exec calls (DELETE + INSERT), got %d", execCount)
	}
}

func TestSyncCommentReactions_SkipsZeroCounts(t *testing.T) {
	var insertCount int
	mq := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			if insertCount == 0 {
				insertCount++ // DELETE
				return pgconn.NewCommandTag("DELETE 1"), nil
			}
			insertCount++
			return pgconn.NewCommandTag("INSERT 0 1"), nil
		},
	}

	err := SyncCommentReactions(context.Background(), mq, 1, map[string]int{"👍": 5, "👎": 0, "❤️": 3})
	if err != nil {
		t.Errorf("SyncCommentReactions: %v", err)
	}
	// DELETE + 2 INSERTs (skipping 👎:0)
	if insertCount != 3 {
		t.Errorf("expected 3 exec calls, got %d", insertCount)
	}
}

func TestSyncCommentReactions_SortedEmojis(t *testing.T) {
	var insertOrder []string
	mq := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			if len(args) >= 2 {
				if emoji, ok := args[1].(string); ok {
					insertOrder = append(insertOrder, emoji)
				}
			}
			return pgconn.NewCommandTag("INSERT 0 1"), nil
		},
	}

	err := SyncCommentReactions(context.Background(), mq, 1, map[string]int{"👎": 1, "👍": 1, "❤️": 1})
	if err != nil {
		t.Errorf("SyncCommentReactions: %v", err)
	}

	// Should be sorted: ❤️, 👍, 👎
	expected := []string{"❤️", "👍", "👎"}
	if len(insertOrder) != len(expected) {
		t.Errorf("insert order length mismatch: %v", insertOrder)
	} else {
		for i, e := range expected {
			if insertOrder[i] != e {
				t.Errorf("insertOrder[%d] = %q, want %q", i, insertOrder[i], e)
			}
		}
	}
}

func TestSyncCommentReactions_ExecError(t *testing.T) {
	mq := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag(""), errors.New("db error")
		},
	}

	err := SyncCommentReactions(context.Background(), mq, 1, map[string]int{"👍": 1})
	if err == nil {
		t.Error("expected error on db failure")
	}
}

func TestSyncCommentReactions_UndefinedTableError(t *testing.T) {
	mq := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			// Return undefined table error (simulating table doesn't exist)
			return pgconn.NewCommandTag(""), &pgconn.PgError{Code: "42P01"}
		},
	}

	// Should not error when table doesn't exist
	err := SyncCommentReactions(context.Background(), mq, 1, map[string]int{"👍": 1})
	if err != nil {
		t.Errorf("expected no error for undefined table, got: %v", err)
	}
}