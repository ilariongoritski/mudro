package app

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	mediadb "github.com/goritskimihail/mudro/internal/media"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type rootRow struct {
	MessageID int64
	Date      time.Time
	Text      string
}

type postState struct {
	ID             int64
	SourcePostID   string
	PublishedAt    time.Time
	Text           string
	Media          json.RawMessage
	LikesCount     int
	ViewsCount     *int
	CommentsCount  *int
	ActualComments int
	HasMediaLinks  bool
	HasReactions   bool
}

type rootMergePlan struct {
	GeneratedAt             string           `json:"generated_at"`
	MaxKnownChannelSourceID int64            `json:"max_known_channel_source_id"`
	TotalTGPosts            int              `json:"total_tg_posts"`
	DiscussionRootPosts     int              `json:"discussion_root_posts"`
	MatchedGroups           int              `json:"matched_groups"`
	RowsToRemove            int              `json:"rows_to_remove"`
	MovedComments           int              `json:"moved_comments"`
	SkippedAmbiguous        int              `json:"skipped_ambiguous"`
	SkippedNoMatch          int              `json:"skipped_no_match"`
	SkippedNoRootSource     int              `json:"skipped_no_root_source"`
	Merges                  []rootMergeGroup `json:"merges"`
	SkippedSample           []rootMergeSkip  `json:"skipped_sample"`
}

type rootMergeGroup struct {
	Reason             string    `json:"reason"`
	RootMessageID      int64     `json:"root_message_id"`
	CanonicalID        int64     `json:"canonical_id"`
	CanonicalSourceID  string    `json:"canonical_source_post_id"`
	DuplicateID        int64     `json:"duplicate_id"`
	DuplicateSourceID  string    `json:"duplicate_source_post_id"`
	DeltaSeconds       int64     `json:"delta_seconds"`
	TextLen            int       `json:"text_len"`
	CommentsMoved      int       `json:"comments_moved"`
	CanonicalPublished time.Time `json:"canonical_published_at"`
	DuplicatePublished time.Time `json:"duplicate_published_at"`
}

type rootMergeSkip struct {
	SourcePostID   string `json:"source_post_id"`
	Reason         string `json:"reason"`
	TextLen        int    `json:"text_len"`
	ActualComments int    `json:"actual_comments"`
}

func Run() {
	apply := flag.Bool("apply", false, "apply root-post merge changes")
	dsn := flag.String("dsn", config.DSN(), "postgres dsn")
	commentsCSV := flag.String("comments-csv", filepath.Join(config.RepoRoot(), "output", "import", "tg_comments_discussion.csv"), "path to discussion CSV export")
	postsCSV := flag.String("posts-csv", filepath.Join(config.RepoRoot(), "output", "import", "tg_posts_mudro.csv"), "path to channel posts CSV export")
	channelJSON := flag.String("channel-json", filepath.Join(config.RepoRoot(), "data", "tg-export", "result.json"), "path to channel result.json export")
	outPath := flag.String("out", "", "path to write merge plan json")
	flag.Parse()

	rootRows, err := readThreadRoots(*commentsCSV)
	if err != nil {
		log.Fatalf("read thread roots: %v", err)
	}
	maxKnownID, err := loadMaxKnownChannelID(*postsCSV, *channelJSON)
	if err != nil {
		log.Fatalf("load max known channel id: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	posts, err := loadTGPosts(ctx, pool)
	if err != nil {
		log.Fatalf("load tg posts: %v", err)
	}

	plan := buildPlan(posts, rootRows, maxKnownID)
	targetOut := strings.TrimSpace(*outPath)
	if targetOut == "" {
		targetOut = filepath.Join(config.RepoRoot(), "output", "db", "tg-root-merge-plan.json")
	}
	if err := os.MkdirAll(filepath.Dir(targetOut), 0o755); err != nil {
		log.Fatalf("mkdir output dir: %v", err)
	}
	if err := writePlan(targetOut, plan); err != nil {
		log.Fatalf("write plan: %v", err)
	}

	log.Printf("PLAN: total_tg_posts=%d discussion_root_posts=%d matched_groups=%d rows_to_remove=%d moved_comments=%d skipped_ambiguous=%d skipped_no_match=%d skipped_no_root_source=%d out=%s",
		plan.TotalTGPosts, plan.DiscussionRootPosts, plan.MatchedGroups, plan.RowsToRemove, plan.MovedComments, plan.SkippedAmbiguous, plan.SkippedNoMatch, plan.SkippedNoRootSource, targetOut)

	if !*apply {
		return
	}

	if err := applyPlan(ctx, pool, plan); err != nil {
		log.Fatalf("apply plan: %v", err)
	}

	log.Printf("DONE: matched_groups=%d removed_rows=%d moved_comments=%d", plan.MatchedGroups, plan.RowsToRemove, plan.MovedComments)
}

func readThreadRoots(path string) (map[int64]rootRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	index := make(map[string]int, len(header))
	for i, name := range header {
		index[strings.TrimSpace(strings.TrimPrefix(name, "\uFEFF"))] = i
	}

	out := make(map[int64]rootRow, 1024)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if field(record, index, "sender_id") != "-1" || field(record, index, "reply_to_msg_id") != "" {
			continue
		}
		id, ok := parseInt64(field(record, index, "message_id"))
		if !ok {
			continue
		}
		date, err := parseCSVDate(field(record, index, "date"))
		if err != nil {
			return nil, fmt.Errorf("parse root date %d: %w", id, err)
		}
		out[id] = rootRow{
			MessageID: id,
			Date:      date.UTC(),
			Text:      normalizeLookupText(normalizeCSVMessage(field(record, index, "message"))),
		}
	}
	return out, nil
}

func loadMaxKnownChannelID(postsCSVPath, channelJSONPath string) (int64, error) {
	maxID := int64(0)

	if value, err := loadMaxIDFromPostsCSV(postsCSVPath); err == nil && value > maxID {
		maxID = value
	} else if err != nil {
		return 0, err
	}

	if value, err := loadMaxIDFromChannelJSON(channelJSONPath); err == nil && value > maxID {
		maxID = value
	} else if err != nil {
		return 0, err
	}

	if maxID == 0 {
		return 0, fmt.Errorf("max known channel id resolved to 0")
	}
	return maxID, nil
}

func loadMaxIDFromPostsCSV(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return 0, err
	}
	index := make(map[string]int, len(header))
	for i, name := range header {
		index[strings.TrimSpace(strings.TrimPrefix(name, "\uFEFF"))] = i
	}

	var maxID int64
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, err
		}
		if id, ok := parseInt64(field(record, index, "message_id")); ok && id > maxID {
			maxID = id
		}
	}
	return maxID, nil
}

type channelExport struct {
	Messages []struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	} `json:"messages"`
}

func loadMaxIDFromChannelJSON(path string) (int64, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var data channelExport
	if err := json.Unmarshal(raw, &data); err != nil {
		return 0, err
	}
	var maxID int64
	for _, msg := range data.Messages {
		if msg.Type != "message" {
			continue
		}
		if msg.ID > maxID {
			maxID = msg.ID
		}
	}
	return maxID, nil
}

func loadTGPosts(ctx context.Context, pool *pgxpool.Pool) ([]postState, error) {
	rows, err := pool.Query(ctx, `
select id, source_post_id, published_at, coalesce(text, ''), media, likes_count, views_count, comments_count
from posts
where source = 'tg'
order by published_at asc, id asc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]postState, 0, 2048)
	for rows.Next() {
		var (
			item       postState
			mediaBytes []byte
		)
		if err := rows.Scan(&item.ID, &item.SourcePostID, &item.PublishedAt, &item.Text, &mediaBytes, &item.LikesCount, &item.ViewsCount, &item.CommentsCount); err != nil {
			return nil, err
		}
		if len(mediaBytes) > 0 {
			item.Media = json.RawMessage(mediaBytes)
		}
		if err := pool.QueryRow(ctx, `select count(*)::int from post_comments where post_id = $1`, item.ID).Scan(&item.ActualComments); err != nil {
			return nil, err
		}
		if err := pool.QueryRow(ctx, `select exists(select 1 from post_media_links where post_id = $1)`, item.ID).Scan(&item.HasMediaLinks); err != nil {
			return nil, err
		}
		if err := pool.QueryRow(ctx, `select exists(select 1 from post_reactions where post_id = $1)`, item.ID).Scan(&item.HasReactions); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func buildPlan(posts []postState, roots map[int64]rootRow, maxKnownID int64) rootMergePlan {
	plan := rootMergePlan{
		GeneratedAt:             time.Now().UTC().Format(time.RFC3339),
		MaxKnownChannelSourceID: maxKnownID,
		TotalTGPosts:            len(posts),
	}

	channel := make([]postState, 0, len(posts))
	extra := make([]postState, 0, 512)
	for _, post := range posts {
		sourceID, ok := parseInt64(post.SourcePostID)
		if !ok {
			channel = append(channel, post)
			continue
		}
		if sourceID > maxKnownID {
			extra = append(extra, post)
			continue
		}
		channel = append(channel, post)
	}
	plan.DiscussionRootPosts = len(extra)

	for _, duplicate := range extra {
		rootID, ok := parseInt64(duplicate.SourcePostID)
		if !ok {
			plan.SkippedNoRootSource++
			plan.addSkipped(duplicate, "invalid-source-post-id")
			continue
		}
		root, ok := roots[rootID]
		if !ok {
			plan.SkippedNoRootSource++
			plan.addSkipped(duplicate, "missing-thread-root-row")
			continue
		}

		bestIdx := -1
		bestDelta := int64(1<<62 - 1)
		ambiguous := false
		for i, candidate := range channel {
			if root.Text != "" && normalizeLookupText(candidate.Text) != root.Text {
				continue
			}
			delta := absDurationSeconds(root.Date.Sub(candidate.PublishedAt))
			if root.Text != "" {
				if delta > int64(6*60*60) {
					continue
				}
			} else if delta > int64(30*60) {
				continue
			}
			if bestIdx == -1 || delta < bestDelta {
				bestIdx = i
				bestDelta = delta
				ambiguous = false
				continue
			}
			if delta == bestDelta {
				ambiguous = true
			}
		}

		if bestIdx == -1 {
			if idx, ok := findComplementCandidate(duplicate, channel); ok {
				canonical := channel[idx]
				plan.addMerge("comments-count-complement", canonical, duplicate, root, absDurationSeconds(root.Date.Sub(canonical.PublishedAt)))
				continue
			}
			plan.SkippedNoMatch++
			plan.addSkipped(duplicate, "no-channel-match")
			continue
		}
		if ambiguous {
			if idx, ok := findComplementCandidate(duplicate, channel); ok {
				canonical := channel[idx]
				plan.addMerge("comments-count-complement", canonical, duplicate, root, absDurationSeconds(root.Date.Sub(canonical.PublishedAt)))
				continue
			}
			plan.SkippedAmbiguous++
			plan.addSkipped(duplicate, "ambiguous-equal-best-match")
			continue
		}

		canonical := channel[bestIdx]
		plan.addMerge("thread-root-match", canonical, duplicate, root, bestDelta)
	}

	sort.Slice(plan.Merges, func(i, j int) bool {
		return plan.Merges[i].DuplicateID < plan.Merges[j].DuplicateID
	})
	return plan
}

func (p *rootMergePlan) addMerge(reason string, canonical, duplicate postState, root rootRow, delta int64) {
	p.MatchedGroups++
	p.RowsToRemove++
	p.MovedComments += duplicate.ActualComments
	p.Merges = append(p.Merges, rootMergeGroup{
		Reason:             reason,
		RootMessageID:      root.MessageID,
		CanonicalID:        canonical.ID,
		CanonicalSourceID:  canonical.SourcePostID,
		DuplicateID:        duplicate.ID,
		DuplicateSourceID:  duplicate.SourcePostID,
		DeltaSeconds:       delta,
		TextLen:            len(strings.TrimSpace(duplicate.Text)),
		CommentsMoved:      duplicate.ActualComments,
		CanonicalPublished: canonical.PublishedAt,
		DuplicatePublished: duplicate.PublishedAt,
	})
}

func (p *rootMergePlan) addSkipped(post postState, reason string) {
	if len(p.SkippedSample) >= 25 {
		return
	}
	p.SkippedSample = append(p.SkippedSample, rootMergeSkip{
		SourcePostID:   post.SourcePostID,
		Reason:         reason,
		TextLen:        len(strings.TrimSpace(post.Text)),
		ActualComments: post.ActualComments,
	})
}

func findComplementCandidate(duplicate postState, channel []postState) (int, bool) {
	if duplicate.ActualComments <= 0 || intValue(duplicate.CommentsCount) != 0 {
		return -1, false
	}
	duplicateText := normalizeLookupText(duplicate.Text)

	bestIdx := -1
	for i, candidate := range channel {
		if candidate.ActualComments != 0 || intValue(candidate.CommentsCount) != duplicate.ActualComments {
			continue
		}
		if absDurationSeconds(candidate.PublishedAt.Sub(duplicate.PublishedAt)) > 5 {
			continue
		}
		if !textCompatible(duplicateText, normalizeLookupText(candidate.Text)) {
			continue
		}
		if bestIdx != -1 {
			return -1, false
		}
		bestIdx = i
	}
	return bestIdx, bestIdx != -1
}

func textCompatible(left, right string) bool {
	if left == "" || right == "" {
		return true
	}
	return left == right || strings.Contains(left, right) || strings.Contains(right, left)
}

func applyPlan(ctx context.Context, pool *pgxpool.Pool, plan rootMergePlan) error {
	for _, merge := range plan.Merges {
		if err := mergePair(ctx, pool, merge.CanonicalID, merge.DuplicateID); err != nil {
			return fmt.Errorf("merge canonical=%d duplicate=%d: %w", merge.CanonicalID, merge.DuplicateID, err)
		}
	}
	return nil
}

func mergePair(ctx context.Context, pool *pgxpool.Pool, canonicalID, duplicateID int64) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	canonical, err := loadPostForUpdate(ctx, tx, canonicalID)
	if err != nil {
		return err
	}
	duplicate, err := loadPostForUpdate(ctx, tx, duplicateID)
	if err != nil {
		return err
	}

	if err := mergePostReactions(ctx, tx, canonicalID, duplicateID); err != nil {
		return err
	}
	if err := movePostComments(ctx, tx, canonicalID, duplicateID); err != nil {
		return err
	}
	if err := mergePostMediaLinks(ctx, tx, canonicalID, duplicateID); err != nil {
		return err
	}

	actualComments, err := countPostComments(ctx, tx, canonicalID)
	if err != nil {
		return err
	}
	mergedMedia, err := mediadb.LoadPostMediaJSON(ctx, tx, []int64{canonicalID})
	if err != nil {
		return err
	}

	mergedText := preferredText(canonical.Text, duplicate.Text)
	mergedLikes := maxInt(canonical.LikesCount, duplicate.LikesCount)
	mergedViews := maxIntPtr(canonical.ViewsCount, duplicate.ViewsCount)
	mergedCommentsCount := maxInt(intValue(canonical.CommentsCount), intValue(duplicate.CommentsCount), actualComments)
	var mergedMediaJSON any
	if raw, ok := mergedMedia[canonicalID]; ok && len(raw) > 0 {
		mergedMediaJSON = raw
	} else if len(canonical.Media) > 0 {
		mergedMediaJSON = canonical.Media
	} else if len(duplicate.Media) > 0 {
		mergedMediaJSON = duplicate.Media
	}

	if _, err := tx.Exec(ctx, `
update posts
set text = $2,
    media = $3,
    likes_count = $4,
    views_count = $5,
    comments_count = $6,
    updated_at = now()
where id = $1
`, canonicalID, nullIfEmpty(mergedText), mergedMediaJSON, mergedLikes, mergedViews, mergedCommentsCount); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `delete from posts where id = $1`, duplicateID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func loadPostForUpdate(ctx context.Context, tx pgx.Tx, postID int64) (postState, error) {
	var (
		item       postState
		mediaBytes []byte
	)
	err := tx.QueryRow(ctx, `
select id, source_post_id, published_at, coalesce(text, ''), media, likes_count, views_count, comments_count
from posts
where id = $1
for update
`, postID).Scan(&item.ID, &item.SourcePostID, &item.PublishedAt, &item.Text, &mediaBytes, &item.LikesCount, &item.ViewsCount, &item.CommentsCount)
	if len(mediaBytes) > 0 {
		item.Media = json.RawMessage(mediaBytes)
	}
	return item, err
}

func mergePostReactions(ctx context.Context, tx pgx.Tx, canonicalID, duplicateID int64) error {
	rows, err := tx.Query(ctx, `
select emoji, max(count) as count
from post_reactions
where post_id = any($1)
group by emoji
order by emoji asc
`, []int64{canonicalID, duplicateID})
	if err != nil {
		return err
	}
	type reactionRow struct {
		emoji string
		count int
	}
	var merged []reactionRow
	for rows.Next() {
		var item reactionRow
		if err := rows.Scan(&item.emoji, &item.count); err != nil {
			rows.Close()
			return err
		}
		merged = append(merged, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, item := range merged {
		if _, err := tx.Exec(ctx, `
insert into post_reactions (post_id, emoji, count)
values ($1, $2, $3)
on conflict (post_id, emoji) do update set count = greatest(post_reactions.count, excluded.count)
`, canonicalID, item.emoji, item.count); err != nil {
			return err
		}
	}
	return nil
}

func movePostComments(ctx context.Context, tx pgx.Tx, canonicalID, duplicateID int64) error {
	_, err := tx.Exec(ctx, `update post_comments set post_id = $1, updated_at = now() where post_id = $2`, canonicalID, duplicateID)
	return err
}

func mergePostMediaLinks(ctx context.Context, tx pgx.Tx, canonicalID, duplicateID int64) error {
	canonicalAssets := make(map[int64]struct{})
	rows, err := tx.Query(ctx, `select media_asset_id from post_media_links where post_id = $1 order by position asc`, canonicalID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var assetID int64
		if err := rows.Scan(&assetID); err != nil {
			rows.Close()
			return err
		}
		canonicalAssets[assetID] = struct{}{}
	}
	rows.Close()

	var nextPosition int
	if err := tx.QueryRow(ctx, `select coalesce(max(position), 0) from post_media_links where post_id = $1`, canonicalID).Scan(&nextPosition); err != nil {
		return err
	}
	nextPosition++

	rows, err = tx.Query(ctx, `select media_asset_id from post_media_links where post_id = $1 order by position asc, media_asset_id asc`, duplicateID)
	if err != nil {
		return err
	}
	var duplicateAssets []int64
	for rows.Next() {
		var assetID int64
		if err := rows.Scan(&assetID); err != nil {
			rows.Close()
			return err
		}
		duplicateAssets = append(duplicateAssets, assetID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, assetID := range duplicateAssets {
		if _, exists := canonicalAssets[assetID]; exists {
			continue
		}
		if _, err := tx.Exec(ctx, `
insert into post_media_links (post_id, media_asset_id, position)
values ($1, $2, $3)
on conflict do nothing
`, canonicalID, assetID, nextPosition); err != nil {
			return err
		}
		canonicalAssets[assetID] = struct{}{}
		nextPosition++
	}
	return nil
}

func countPostComments(ctx context.Context, tx pgx.Tx, postID int64) (int, error) {
	var out int
	err := tx.QueryRow(ctx, `select count(*)::int from post_comments where post_id = $1`, postID).Scan(&out)
	return out, err
}

func writePlan(path string, plan rootMergePlan) error {
	payload, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func field(record []string, index map[string]int, key string) string {
	position, ok := index[key]
	if !ok || position >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[position])
}

func parseCSVDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05-07:00",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, raw); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format")
}

func normalizeCSVMessage(raw string) string {
	text := strings.TrimSpace(strings.ReplaceAll(raw, "\r\n", "\n"))
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	if text == "" || text == "MediaMessage" {
		return ""
	}
	return text
}

func normalizeLookupText(raw string) string {
	text := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(raw, "\r\n", "\n")))
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return text
}

func parseInt64(raw string) (int64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	return v, err == nil
}

func absDurationSeconds(d time.Duration) int64 {
	if d < 0 {
		d = -d
	}
	return int64(d / time.Second)
}

func preferredText(left, right string) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if len(right) > len(left) {
		return right
	}
	return left
}

func nullIfEmpty(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return raw
}

func intValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func maxInt(values ...int) int {
	best := values[0]
	for _, value := range values[1:] {
		if value > best {
			best = value
		}
	}
	return best
}

func maxIntPtr(left, right *int) any {
	best := intValue(left)
	if value := intValue(right); value > best {
		best = value
	}
	if left == nil && right == nil {
		return nil
	}
	return best
}
