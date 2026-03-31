package feed

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/goritskimihail/mudro/internal/posts"
)

// handlePosts returns a paginated list of posts.
func (s *Server) handlePosts(w http.ResponseWriter, r *http.Request) {
	if s.postsSvc == nil {
		http.Error(w, "posts service unavailable", http.StatusServiceUnavailable)
		return
	}
	ctx := r.Context()

	limit := parseLimit(r.URL.Query().Get("limit"))
	page, err := parsePage(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "invalid page: "+err.Error(), http.StatusBadRequest)
		return
	}
	cursorTS, cursorID, err := parseCursor(r.URL.Query().Get("before_ts"), r.URL.Query().Get("before_id"))
	if err != nil {
		http.Error(w, "invalid cursor: "+err.Error(), http.StatusBadRequest)
		return
	}
	source, err := parseSource(r.URL.Query().Get("source"))
	if err != nil {
		http.Error(w, "invalid source: "+err.Error(), http.StatusBadRequest)
		return
	}
	sortOrder, err := parseSort(r.URL.Query().Get("sort"))
	if err != nil {
		http.Error(w, "invalid sort: "+err.Error(), http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	items, next, err := s.postsSvc.LoadPosts(ctx, cursorTS, cursorID, page, limit, source, posts.SortOrder(sortOrder), query)
	if err != nil {
		log.Printf("loadPosts: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := postsResponse{Page: page, Limit: limit, Items: items, NextCursor: next}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}

// handleFront returns the API response for the frontend feed.
// P0: uses postsSvc.LoadLastSyncAt — no direct pool access.
func (s *Server) handleFront(w http.ResponseWriter, r *http.Request) {
	if s.postsSvc == nil {
		http.Error(w, "posts service unavailable", http.StatusServiceUnavailable)
		return
	}
	ctx := r.Context()
	limit := parseLimit(r.URL.Query().Get("limit"))
	source, err := parseSource(r.URL.Query().Get("source"))
	if err != nil {
		http.Error(w, "invalid source: "+err.Error(), http.StatusBadRequest)
		return
	}
	sortOrder, err := parseSort(r.URL.Query().Get("sort"))
	if err != nil {
		http.Error(w, "invalid sort: "+err.Error(), http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	items, next, err := s.postsSvc.LoadPosts(ctx, nil, nil, nil, limit, source, posts.SortOrder(sortOrder), query)
	if err != nil {
		log.Printf("loadPosts(front): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	totalPosts, err := s.postsSvc.CountVisiblePosts(ctx)
	if err != nil {
		log.Printf("front count posts: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// P0 fix: via service, not direct pool
	lastSync, err := s.postsSvc.LoadLastSyncAt(ctx)
	if err != nil {
		log.Printf("front max(updated_at): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sourceStats, err := s.postsSvc.LoadSourceStats(ctx)
	if err != nil {
		log.Printf("front source stats: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := frontResponse{
		Meta: frontMeta{TotalPosts: totalPosts, LastSyncAt: lastSync, Sources: sourceStats},
		Feed: postsResponse{Limit: limit, Items: items, NextCursor: next},
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}

// handleFeed renders the HTML debug feed page.
func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, _, err := parseCursor(r.URL.Query().Get("before_ts"), r.URL.Query().Get("before_id")); err != nil {
		http.Error(w, "invalid cursor: "+err.Error(), http.StatusBadRequest)
		return
	}
	limit := parseLimit(r.URL.Query().Get("limit"))
	page, err := parsePage(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "invalid page: "+err.Error(), http.StatusBadRequest)
		return
	}
	if page == nil {
		defaultPage := 1
		page = &defaultPage
	}
	source, err := parseSource(r.URL.Query().Get("source"))
	if err != nil {
		http.Error(w, "invalid source: "+err.Error(), http.StatusBadRequest)
		return
	}
	sortOrder, err := parseSort(r.URL.Query().Get("sort"))
	if err != nil {
		http.Error(w, "invalid sort: "+err.Error(), http.StatusBadRequest)
		return
	}

	q := r.URL.Query().Get("q")
	items, _, err := s.postsSvc.LoadPosts(ctx, nil, nil, page, limit, source, posts.SortOrder(sortOrder), q)
	if err != nil {
		log.Printf("loadPosts(feed): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var totalPosts int64
	if err := s.countVisiblePosts(ctx, &totalPosts); err != nil {
		log.Printf("feed count posts: %v", err)
	}
	sourceStats, err := s.loadSourceStats(ctx)
	if err != nil {
		log.Printf("feed source stats: %v", err)
	}
	vkTotal, tgTotal := sourceTotals(sourceStats)

	data := feedPageData{
		Limit: limit, Page: *page, Source: source, SortOrder: sortOrder,
		AllURL: buildFeedURL(limit, 1, "", sortOrder), VKURL: buildFeedURL(limit, 1, "vk", sortOrder),
		TGURL: buildFeedURL(limit, 1, "tg", sortOrder), NewestURL: buildFeedURL(limit, 1, source, "desc"),
		OldestURL: buildFeedURL(limit, 1, source, "asc"), Items: make([]feedItem, 0, len(items)),
		SourceName: sourceLabel(source), TotalPosts: totalPosts, VKTotal: vkTotal, TGTotal: tgTotal,
	}

	for _, p := range items {
		originalURL := buildOriginalPostURL(p.Source, p.SourcePostID)
		commentsURL := ""
		if p.CommentsCount != nil && *p.CommentsCount > 0 && originalURL != "" {
			commentsURL = originalURL
		}
		data.Items = append(data.Items, feedItem{
			ID: p.ID, Source: p.Source, SourcePostID: p.SourcePostID,
			PublishedAt: p.PublishedAt.Format("2006-01-02 15:04:05"),
			Text: compactText(p.Text), LikesCount: p.LikesCount,
			ViewsCount: p.ViewsCount, CommentsCount: p.CommentsCount,
			OriginalURL: originalURL, CommentsURL: commentsURL,
			Reactions: buildFeedReactions(p.Reactions),
			Comments:  buildFeedComments(p.Comments),
			Media:     parseMediaItems(p.Media),
		})
	}
	if len(items) == limit {
		data.NextURL = buildFeedURL(limit, *page+1, source, sortOrder)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := feedPageTmpl.Execute(w, data); err != nil {
		log.Printf("template feed: %v", err)
	}
}
