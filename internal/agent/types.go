package agent

import "time"

const (
	StatusQueued          = "queued"
	StatusWaitingApproval = "waiting_approval"
	StatusInProgress      = "in_progress"
	StatusDone            = "done"
	StatusFailed          = "failed"
	StatusRejected        = "rejected"
)

type Task struct {
	ID          int64
	Kind        string
	Payload     []byte
	Status      string
	Priority    int
	Attempts    int
	MaxAttempts int
	DedupeKey   string
	RunAfter    time.Time
	LockedBy    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
