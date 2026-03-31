package feed

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func parseLimit(raw string) int {
	const (
		def = 50
		max = 200
	)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

func parsePage(raw string) (*int, error) {
	if raw == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return nil, errors.New("page must be a positive integer")
	}
	return &n, nil
}

func parseCursor(tsRaw, idRaw string) (*time.Time, *int64, error) {
	if tsRaw == "" && idRaw == "" {
		return nil, nil, nil
	}
	if tsRaw == "" || idRaw == "" {
		return nil, nil, errors.New("both before_ts and before_id are required")
	}
	ts, err := time.Parse(time.RFC3339, tsRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("before_ts: %w", err)
	}
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("before_id: %w", err)
	}
	return &ts, &id, nil
}

func parseSource(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "all":
		return "", nil
	case "vk":
		return "vk", nil
	case "tg":
		return "tg", nil
	default:
		return "", errors.New("use all|vk|tg")
	}
}

func sourceLabel(source string) string {
	if source == "" {
		return "all"
	}
	return source
}

func parseSort(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "desc":
		return "desc", nil
	case "asc":
		return "asc", nil
	default:
		return "", errors.New("use asc|desc")
	}
}

func buildFeedURL(limit, page int, source, sortOrder string) string {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("page", strconv.Itoa(page))
	if source != "" {
		q.Set("source", source)
	}
	if sortOrder != "" {
		q.Set("sort", sortOrder)
	}
	return "/feed?" + q.Encode()
}

func buildOriginalPostURL(source, sourcePostID string) string {
	switch source {
	case "vk":
		if strings.Contains(sourcePostID, "_") {
			return "https://vk.com/wall" + sourcePostID
		}
	}
	return ""
}
