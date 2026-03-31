package usecase

import (
	"context"
	"time"

	"github.com/goritskimihail/mudro/internal/agent/domain"
	"github.com/goritskimihail/mudro/internal/agent/repository"
)

type TaskUsecase interface {
	ClaimNext(ctx context.Context, workerID string) (*domain.Task, error)
	CompleteTask(ctx context.Context, taskID int64) error
	FailTask(ctx context.Context, taskID int64, errText string, retryAfter time.Duration) error
	
	Enqueue(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error)
	EnqueueWaitingApproval(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error)
	
	ApproveTask(ctx context.Context, taskID int64) error
	RejectTask(ctx context.Context, taskID int64, reason string) error
}

type agentService struct {
	repo repository.TaskRepository
}

func NewAgentService(repo repository.TaskRepository) TaskUsecase {
	return &agentService{repo: repo}
}

func (s *agentService) ClaimNext(ctx context.Context, workerID string) (*domain.Task, error) {
	return s.repo.ClaimNext(ctx, workerID)
}

func (s *agentService) CompleteTask(ctx context.Context, taskID int64) error {
	return s.repo.MarkDone(ctx, taskID)
}

func (s *agentService) FailTask(ctx context.Context, taskID int64, errText string, retryAfter time.Duration) error {
	return s.repo.MarkFailed(ctx, taskID, errText, retryAfter)
}

func (s *agentService) Enqueue(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
	return s.repo.Enqueue(ctx, kind, payload, priority, runAfter, maxAttempts, dedupeKey)
}

func (s *agentService) EnqueueWaitingApproval(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
	return s.repo.EnqueueWaitingApproval(ctx, kind, payload, priority, runAfter, maxAttempts, dedupeKey)
}

func (s *agentService) ApproveTask(ctx context.Context, taskID int64) error {
	return s.repo.ApproveTask(ctx, taskID)
}

func (s *agentService) RejectTask(ctx context.Context, taskID int64, reason string) error {
	return s.repo.RejectTask(ctx, taskID, reason)
}
