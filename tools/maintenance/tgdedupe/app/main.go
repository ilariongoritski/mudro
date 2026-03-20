package app

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	mediadb "github.com/goritskimihail/mudro/internal/media"
	"github.com/goritskimihail/mudro/internal/tgexport"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	Visible        bool
}

type dedupePlan struct {
	GeneratedAt        string          `json:"generated_at"`
	TotalTGPosts       int             `json:"total_tg_posts"`
	GroupsToMerge      int             `json:"groups_to_merge"`
	RowsToRemove       int             `json:"rows_to_remove"`
	SkippedGroups      int             `json:"skipped_groups"`
	MergedGroups       []dedupeGroup   `json:"merged_groups"`
	SkippedGroupSample []dedupePreview `json:"skipped_group_sample"`
}

type dedupeGroup struct {
	Reason       string          `json:"reason"`
	CanonicalID  int64           `json:"canonical_id"`
	CanonicalSP  string          `json:"canonical_source_post_id"`
	DuplicateIDs []int64         `json:"duplicate_ids"`
	Posts        []dedupePreview `json:"posts"`
}

type dedupePreview struct {
	ID             int64     `json:"id"`
	SourcePostID   string    `json:"source_post_id"`
	PublishedAt    time.Time `json:"published_at"`
	TextLen        int       `json:"text_len"`
	HasMediaLinks  bool      `json:"has_media_links"`
	HasReactions   bool      `json:"has_reactions"`
	CommentsCount  *int      `json:"comments_count,omitempty"`
	ActualComments int       `json:"actual_comments"`
	Visible        bool      `json:"visible"`
}

func Run() {
	apply := flag.Bool("apply", false, "apply dedupe changes")
	dsn := flag.String("dsn", config.DSN(), "postgres dsn")
	outPath := flag.String("out", "", "path to write dedupe plan json")
	flag.Parse()

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

	visibleIDs, _, err := tgexport.LoadVisibleSourcePostIDsFromRepo(config.RepoRoot())
	if err != nil {
		log.Printf("tg visibility ids unavailable, continuing without them: %v", err)
	}
	visibleSet := make(map[string]struct{}, len(visibleIDs))
	for _, id := range visibleIDs {
		visibleSet[id] = struct{}{}
	}

	posts, err := loadTGPosts(ctx, pool, visibleSet)
	if err != nil {
		log.Fatalf("load tg posts: %v", err)
	}

	plan := buildPlan(posts)

	targetOut := strings.TrimSpace(*outPath)
	if targetOut == "" {
		targetOut = filepath.Join(config.RepoRoot(), "output", "db", "tg-dedupe-plan.json")
	}
	if err := os.MkdirAll(filepath.Dir(targetOut), 0o755); err != nil {
		log.Fatalf("mkdir output dir: %v", err)
	}
	if err := writePlan(targetOut, plan); err != nil {
		log.Fatalf("write plan: %v", err)
	}

	log.Printf("PLAN: total_tg_posts=%d groups_to_merge=%d rows_to_remove=%d skipped_groups=%d out=%s",
		plan.TotalTGPosts, plan.GroupsToMerge, plan.RowsToRemove, plan.SkippedGroups, targetOut)

	if !*apply {
		return
	}

	if err := applyPlan(ctx, pool, plan); err != nil {
		log.Fatalf("apply plan: %v", err)
	}

	log.Printf("DONE: merged_groups=%d removed_rows=%d", plan.GroupsToMerge, plan.RowsToRemove)
}

func loadTGPosts(ctx context.Context, pool *pgxpool.Pool, visibleSet map[string]struct{}) ([]postState, error) {
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
		_, item.Visible = visibleSet[item.SourcePostID]
		out = append(out, item)
	}
	return out, rows.Err()
}

func buildPlan(posts []postState) dedupePlan {
	groups := make(map[string][]postState)
	for _, post := range posts {
		key := candidateKey(post)
		if key == "" {
			continue
		}
		groups[key] = append(groups[key], post)
	}

	plan := dedupePlan{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		TotalTGPosts: len(posts),
	}

	for _, group := range groups {
		sort.Slice(group, func(i, j int) bool {
			if group[i].PublishedAt.Equal(group[j].PublishedAt) {
				return group[i].ID < group[j].ID
			}
			return group[i].PublishedAt.Before(group[j].PublishedAt)
		})

		reason, safe := classifyGroup(group)
		if !safe {
			plan.SkippedGroups++
			if len(plan.SkippedGroupSample) < 20 {
				plan.SkippedGroupSample = append(plan.SkippedGroupSample, buildPreviewGroup(group)...)
			}
			continue
		}

		sort.Slice(group, func(i, j int) bool {
			return canonicalLess(group[i], group[j])
		})
		canonical := group[0]
		duplicates := make([]int64, 0, len(group)-1)
		for _, item := range group[1:] {
			duplicates = append(duplicates, item.ID)
		}

		plan.GroupsToMerge++
		plan.RowsToRemove += len(duplicates)
		plan.MergedGroups = append(plan.MergedGroups, dedupeGroup{
			Reason:       reason,
			CanonicalID:  canonical.ID,
			CanonicalSP:  canonical.SourcePostID,
			DuplicateIDs: duplicates,
			Posts:        buildPreviewGroup(group),
		})
	}

	sort.Slice(plan.MergedGroups, func(i, j int) bool {
		return plan.MergedGroups[i].CanonicalID < plan.MergedGroups[j].CanonicalID
	})
	return plan
}

func candidateKey(post postState) string {
	text := normalizeLookupText(post.Text)
	if text != "" {
		return fmt.Sprintf("text:%x:%d", md5Bytes(text), post.PublishedAt.Unix()/5)
	}
	return fmt.Sprintf("empty:%d", post.PublishedAt.Unix()/5)
}

func classifyGroup(group []postState) (string, bool) {
	if len(group) != 2 {
		return "ambiguous-multi-group", false
	}

	textA := normalizeLookupText(group[0].Text)
	textB := normalizeLookupText(group[1].Text)
	if textA != textB {
		return "different-text", false
	}
	if textA != "" {
		return "same-text-near-time", true
	}

	if group[0].Visible != group[1].Visible {
		return "empty-text-visible-mismatch", true
	}
	if group[0].HasMediaLinks != group[1].HasMediaLinks {
		return "empty-text-media-mismatch", true
	}
	if group[0].HasReactions != group[1].HasReactions {
		return "empty-text-reaction-mismatch", true
	}
	if group[0].ActualComments != group[1].ActualComments {
		return "empty-text-actual-comments-mismatch", true
	}
	if intValue(group[0].CommentsCount) != intValue(group[1].CommentsCount) {
		return "empty-text-comments-count-mismatch", true
	}

	return "empty-text-identical", false
}

func canonicalLess(left, right postState) bool {
	leftScore := score(left)
	rightScore := score(right)
	if leftScore != rightScore {
		return leftScore > rightScore
	}
	if left.PublishedAt.Equal(right.PublishedAt) {
		return left.ID > right.ID
	}
	return left.PublishedAt.After(right.PublishedAt)
}

func score(item postState) int64 {
	var score int64
	if item.Visible {
		score += 1_000_000
	}
	score += int64(item.ActualComments) * 10_000
	if item.HasMediaLinks {
		score += 1_000
	}
	if item.HasReactions {
		score += 100
	}
	score += int64(intValue(item.CommentsCount)) * 10
	score += int64(item.LikesCount)
	return score
}

func applyPlan(ctx context.Context, pool *pgxpool.Pool, plan dedupePlan) error {
	for _, group := range plan.MergedGroups {
		for _, duplicateID := range group.DuplicateIDs {
			if err := mergePost(ctx, pool, group.CanonicalID, duplicateID); err != nil {
				return fmt.Errorf("merge canonical=%d duplicate=%d: %w", group.CanonicalID, duplicateID, err)
			}
		}
	}
	return nil
}

func mergePost(ctx context.Context, pool *pgxpool.Pool, canonicalID, duplicateID int64) error {
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

func writePlan(path string, plan dedupePlan) error {
	payload, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func buildPreviewGroup(group []postState) []dedupePreview {
	out := make([]dedupePreview, 0, len(group))
	for _, item := range group {
		out = append(out, dedupePreview{
			ID:             item.ID,
			SourcePostID:   item.SourcePostID,
			PublishedAt:    item.PublishedAt,
			TextLen:        len(strings.TrimSpace(item.Text)),
			HasMediaLinks:  item.HasMediaLinks,
			HasReactions:   item.HasReactions,
			CommentsCount:  item.CommentsCount,
			ActualComments: item.ActualComments,
			Visible:        item.Visible,
		})
	}
	return out
}

func normalizeLookupText(raw string) string {
	text := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(raw, "\r\n", "\n")))
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return text
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

func md5Bytes(raw string) [16]byte {
	return md5Sum([]byte(raw))
}
