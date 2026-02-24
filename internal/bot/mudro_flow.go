package bot

import (
	"strings"
	"sync"
	"time"
)

type pendingChoice struct {
	question string
	answer   string
	created  time.Time
}

type mudroFlowStore struct {
	mu     sync.Mutex
	byChat map[int64]pendingChoice
}

var flowStore = mudroFlowStore{
	byChat: make(map[int64]pendingChoice),
}

func savePendingChoice(chatID int64, question string, answer string) {
	flowStore.mu.Lock()
	defer flowStore.mu.Unlock()
	flowStore.byChat[chatID] = pendingChoice{
		question: strings.TrimSpace(question),
		answer:   strings.TrimSpace(answer),
		created:  time.Now(),
	}
}

func loadPendingChoice(chatID int64) (pendingChoice, bool) {
	flowStore.mu.Lock()
	defer flowStore.mu.Unlock()
	v, ok := flowStore.byChat[chatID]
	if !ok {
		return pendingChoice{}, false
	}
	if time.Since(v.created) > 20*time.Minute {
		delete(flowStore.byChat, chatID)
		return pendingChoice{}, false
	}
	return v, true
}

func clearPendingChoice(chatID int64) {
	flowStore.mu.Lock()
	defer flowStore.mu.Unlock()
	delete(flowStore.byChat, chatID)
}

func detectChoicePrompt(answer string) bool {
	lines := strings.Split(strings.ToLower(answer), "\n")
	has1 := false
	has2 := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "1)") || strings.HasPrefix(line, "1 ") {
			has1 = true
		}
		if strings.HasPrefix(line, "2.") || strings.HasPrefix(line, "2)") || strings.HasPrefix(line, "2 ") {
			has2 = true
		}
	}
	return has1 && has2
}

func isChoiceReply(text string) (string, bool) {
	t := strings.TrimSpace(text)
	if t == "1" || t == "2" {
		return t, true
	}
	return "", false
}
