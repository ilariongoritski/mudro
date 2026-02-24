package bot

import (
	"strings"
)

func formatReply(prefix string, body []byte, err error) string {
	if err != nil {
		return trimMessage(prefix+": error: "+err.Error()+"\n"+string(body), NewRunner().Limit)
	}
	if len(body) == 0 {
		return trimMessage(prefix+": ok", NewRunner().Limit)
	}
	return trimMessage(prefix+":\n"+string(body), NewRunner().Limit)
}

func trimMessage(s string, limit int) string {
	if limit <= 0 {
		return s
	}
	if len([]rune(s)) <= limit {
		return s
	}
	suffix := "\n...(truncated)"
	avail := limit - len([]rune(suffix))
	if avail <= 0 {
		return suffix
	}
	r := []rune(s)
	if avail > len(r) {
		avail = len(r)
	}
	return strings.TrimSpace(string(r[:avail])) + suffix
}
