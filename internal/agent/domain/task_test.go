package domain

import (
	"testing"
	"time"
)

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"StatusQueued", StatusQueued},
		{"StatusWaitingApproval", StatusWaitingApproval},
		{"StatusInProgress", StatusInProgress},
		{"StatusDone", StatusDone},
		{"StatusFailed", StatusFailed},
		{"StatusRejected", StatusRejected},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value == "" {
				t.Errorf("%s should not be empty", tc.name)
			}
		})
	}
}

func TestTask_Fields(t *testing.T) {
	now := time.Now()
	runAfter := now.Add(time.Hour)

	task := Task{
		ID:          42,
		Kind:        "test_task",
		Payload:     []byte(`{"key":"value"}`),
		Status:      StatusQueued,
		Priority:    10,
		Attempts:    0,
		MaxAttempts: 3,
		DedupeKey:   "test-dedupe-key",
		RunAfter:    runAfter,
		LockedBy:    "worker-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if task.ID != 42 {
		t.Errorf("ID = %d, want 42", task.ID)
	}
	if task.Kind != "test_task" {
		t.Errorf("Kind = %q, want %q", task.Kind, "test_task")
	}
	if string(task.Payload) != `{"key":"value"}` {
		t.Errorf("Payload = %q, want %q", string(task.Payload), `{"key":"value"}`)
	}
	if task.Status != StatusQueued {
		t.Errorf("Status = %q, want %q", task.Status, StatusQueued)
	}
	if task.Priority != 10 {
		t.Errorf("Priority = %d, want 10", task.Priority)
	}
	if task.Attempts != 0 {
		t.Errorf("Attempts = %d, want 0", task.Attempts)
	}
	if task.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", task.MaxAttempts)
	}
	if task.DedupeKey != "test-dedupe-key" {
		t.Errorf("DedupeKey = %q, want %q", task.DedupeKey, "test-dedupe-key")
	}
	if !task.RunAfter.Equal(runAfter) {
		t.Errorf("RunAfter = %v, want %v", task.RunAfter, runAfter)
	}
	if task.LockedBy != "worker-1" {
		t.Errorf("LockedBy = %q, want %q", task.LockedBy, "worker-1")
	}
	if !task.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", task.CreatedAt, now)
	}
	if !task.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", task.UpdatedAt, now)
	}
}

func TestTask_ZeroValues(t *testing.T) {
	var task Task

	if task.ID != 0 {
		t.Errorf("ID zero value = %d, want 0", task.ID)
	}
	if task.Kind != "" {
		t.Errorf("Kind zero value = %q, want empty", task.Kind)
	}
	if task.Payload != nil {
		t.Errorf("Payload zero value = %v, want nil", task.Payload)
	}
	if task.Status != "" {
		t.Errorf("Status zero value = %q, want empty", task.Status)
	}
	if task.Priority != 0 {
		t.Errorf("Priority zero value = %d, want 0", task.Priority)
	}
	if task.Attempts != 0 {
		t.Errorf("Attempts zero value = %d, want 0", task.Attempts)
	}
	if task.MaxAttempts != 0 {
		t.Errorf("MaxAttempts zero value = %d, want 0", task.MaxAttempts)
	}
	if task.DedupeKey != "" {
		t.Errorf("DedupeKey zero value = %q, want empty", task.DedupeKey)
	}
	if !task.RunAfter.IsZero() {
		t.Errorf("RunAfter zero value = %v, want zero", task.RunAfter)
	}
	if task.LockedBy != "" {
		t.Errorf("LockedBy zero value = %q, want empty", task.LockedBy)
	}
	if !task.CreatedAt.IsZero() {
		t.Errorf("CreatedAt zero value = %v, want zero", task.CreatedAt)
	}
	if !task.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt zero value = %v, want zero", task.UpdatedAt)
	}
}