package repository

import (
	"context"
	"time"

	"github.com/goritskimihail/mudro/internal/agent/domain"
)

type TaskRepository interface {
	Enqueue(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error)
	EnqueueWaitingApproval(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error)
	
	ApproveTask(ctx context.Context, id int64) error
	RejectTask(ctx context.Context, id int64, reason string) error

	ClaimNext(ctx context.Context, worker string) (*domain.Task, error)
	MarkDone(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, errText string, retryAfter time.Duration) error
}
