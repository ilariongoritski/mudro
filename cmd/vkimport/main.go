package main

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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
vkimport: VK RAW -> Postgres (MVP archive snapshot)

- Reads vk_wall_*.json from --dir (default: ~/vk-export)
- Normalizes to posts + media tables
- Idempotent: rerun does not create duplicates (UPSERT by (source, source_post_id))
- 1 post = 1 transaction
- media replaced each run (delete + batch insert)
*/

func main() {
	var (
		dir = flag.String("dir", filepath.Join(os.Getenv("HOME"), "vk-export"), "dir with vk_wall_*.json")
		dsn = flag.String("dsn", envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable"), "postgres dsn")
	)
	flag.Parse()

	// global app ctx: used only for setup; per-tx timeouts are inside repo methods
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(*dir, "vk_wall_*.json"))
	if err != nil {
		log.Fatalf("glob: %v", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		log.Fatalf("no vk_wall_*.json in %s", *dir)
	}

	repo := NewRepo(pool)

	total := 0
	for _, f := range files {
		n, err := importFile(context.Background(), repo, f) // per-post tx timeouts in repo
		if err != nil {
			log.Fatalf("import %s: %v", filepath.Base(f), err)
		}
		total += n
		log.Printf("OK %s: posts=%d", filepath.Base(f), n)
	}

	log.Printf("DONE: total posts=%d", total)
}

func envOr(k, def string) string {
	v := os.Getenv(k)
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func importFile(ctx context.Context, repo *Repo, path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var page VKWallPage
	if err := json.Unmarshal(b, &page); err != nil {
		return 0, fmt.Errorf("unmarshal %s: %w", filepath.Base(path), err)
	}
	if page.Error != nil {
		return 0, fmt.Errorf("vk error %d: %s", page.Error.ErrorCode, page.Error.ErrorMsg)
	}

	rawFile := filepath.Base(path)

	for i, p := range page.Response.Items {
		post := MapVKPostToUnified(p)
		post.RawFile = rawFile
		post.RawIndex = i

		medias := MapVKAttachmentsToMedia(p.Attachments)

		if err := repo.UpsertPostWithMedia(ctx, post, medias); err != nil {
			return 0, fmt.Errorf("post %s[%d] (%d_%d): %w", rawFile, i, p.OwnerID, p.ID, err)
		}
	}

	return len(page.Response.Items), nil
}

//
// ===== VK DTO =====
//

type VKWallPage struct {
	Response struct {
		Items []VKPost `json:"items"`
	} `json:"response"`
	Error *VKError `json:"error,omitempty"`
}

type VKError struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type VKPost struct {
	ID      int64  `json:"id"`
	OwnerID int64  `json:"owner_id"`
	Date    int64  `json:"date"`
	Text    string `json:"text"`

	Likes    VKCount  `json:"likes"`
	Comments VKCount  `json:"comments"`
	Reposts  VKCount  `json:"reposts"`
	Views    *VKCount `json:"views,omitempty"`

	Attachments []VKAttachment `json:"attachments,omitempty"`
}

type VKCount struct {
	Count int `json:"count"`
}

type VKAttachment struct {
	Type  string   `json:"type"`
	Photo *VKPhoto `json:"photo,omitempty"`
	Doc   *VKDoc   `json:"doc,omitempty"`
	Link  *VKLink  `json:"link,omitempty"`
	Video *VKVideo `json:"video,omitempty"`
	Audio *VKAudio `json:"audio,omitempty"`
}

type VKPhoto struct {
	ID    int64         `json:"id"`
	Sizes []VKPhotoSize `json:"sizes"`
}
type VKPhotoSize struct {
	URL string `json:"url"`
	W   int    `json:"width"`
	H   int    `json:"height"`
}

type VKDoc struct {
	ID  int64  `json:"id"`
	Ext string `json:"ext"`
	URL string `json:"url"`
}

type VKLink struct {
	URL string `json:"url"`
}

type VKVideo struct {
	ID      int64 `json:"id"`
	OwnerID int64 `json:"owner_id"`
	Image   []struct {
		URL string `json:"url"`
		W   int    `json:"width"`
		H   int    `json:"height"`
	} `json:"image,omitempty"`
}

type VKAudio struct {
	ID int64 `json:"id"`
}

//
// ===== Contract (MVP) =====
//

type UnifiedPost struct {
	Source        string
	SourcePostID  string
	SourceURL     string
	PublishedAt   time.Time
	Text          string
	Likes         int
	ViewsNullable *int32 // nullable in DB
	CommentsCount int
	Reposts       int

	RawFile  string
	RawIndex int
}

type MediaItem struct {
	Kind       string `json:"kind"`
	URL        string `json:"url"`
	PreviewURL string `json:"preview_url,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Position   int    `json:"position"`
}

func MapVKPostToUnified(p VKPost) UnifiedPost {
	sourcePostID := fmt.Sprintf("%d_%d", p.OwnerID, p.ID)

	var viewsPtr *int32
	if p.Views != nil {
		v := int32(p.Views.Count)
		viewsPtr = &v
	}

	return UnifiedPost{
		Source:        "vk",
		SourcePostID:  sourcePostID,
		SourceURL:     fmt.Sprintf("https://vk.com/wall%d_%d", p.OwnerID, p.ID),
		PublishedAt:   time.Unix(p.Date, 0).UTC(),
		Text:          p.Text,
		Likes:         p.Likes.Count,
		ViewsNullable: viewsPtr,
		CommentsCount: p.Comments.Count,
		Reposts:       p.Reposts.Count,
	}
}

func MapVKAttachmentsToMedia(atts []VKAttachment) []MediaItem {
	out := make([]MediaItem, 0, len(atts))
	pos := 0

	for _, a := range atts {
		switch a.Type {
		case "photo":
			if a.Photo == nil || len(a.Photo.Sizes) == 0 {
				continue
			}
			orig := pickLargest(a.Photo.Sizes)
			if orig.URL == "" {
				continue
			}
			prev := pickPreview(a.Photo.Sizes, 480)
			out = append(out, MediaItem{
				Kind:       "photo",
				URL:        orig.URL,
				PreviewURL: prev.URL,
				Width:      prev.W,
				Height:     prev.H,
				Position:   pos,
			})
			pos++

		case "doc":
			if a.Doc == nil {
				continue
			}
			kind := "doc"
			if strings.ToLower(a.Doc.Ext) == "gif" {
				kind = "gif"
			}
			out = append(out, MediaItem{Kind: kind, URL: a.Doc.URL, Position: pos})
			pos++

		case "link":
			if a.Link == nil {
				continue
			}
			out = append(out, MediaItem{Kind: "link", URL: a.Link.URL, Position: pos})
			pos++

		case "video":
			if a.Video == nil {
				out = append(out, MediaItem{Kind: "video", Position: pos})
				pos++
				continue
			}
			videoURL := fmt.Sprintf("https://vk.com/video%d_%d", a.Video.OwnerID, a.Video.ID)

			prevURL, w, h := "", 0, 0
			if len(a.Video.Image) > 0 {
				best := pickPreviewGeneric(a.Video.Image, 480)
				prevURL, w, h = best.URL, best.W, best.H
			}

			out = append(out, MediaItem{
				Kind:       "video",
				URL:        videoURL,
				PreviewURL: prevURL,
				Width:      w,
				Height:     h,
				Position:   pos,
			})
			pos++

		case "audio":
			out = append(out, MediaItem{Kind: "audio", Position: pos})
			pos++
		}
	}

	return out
}

func pickLargest(sizes []VKPhotoSize) VKPhotoSize {
	best := VKPhotoSize{}
	bestArea := -1
	for _, s := range sizes {
		if s.URL == "" || s.W <= 0 || s.H <= 0 {
			continue
		}
		area := s.W * s.H
		if area > bestArea {
			bestArea = area
			best = s
		}
	}
	return best
}

func pickPreview(sizes []VKPhotoSize, targetW int) VKPhotoSize {
	var cand []VKPhotoSize
	for _, s := range sizes {
		if s.URL != "" && s.W > 0 && s.H > 0 {
			cand = append(cand, s)
		}
	}
	if len(cand) == 0 {
		return VKPhotoSize{}
	}
	sort.Slice(cand, func(i, j int) bool { return cand[i].W < cand[j].W })
	for _, s := range cand {
		if s.W >= targetW {
			return s
		}
	}
	return cand[len(cand)-1]
}

func pickPreviewGeneric(imgs []struct {
	URL string `json:"url"`
	W   int    `json:"width"`
	H   int    `json:"height"`
}, targetW int) (best struct {
	URL string
	W   int
	H   int
}) {
	if len(imgs) == 0 {
		return best
	}
	type it struct {
		URL  string
		W, H int
	}
	cand := make([]it, 0, len(imgs))
	for _, s := range imgs {
		if s.URL != "" && s.W > 0 && s.H > 0 {
			cand = append(cand, it{s.URL, s.W, s.H})
		}
	}
	if len(cand) == 0 {
		return best
	}
	sort.Slice(cand, func(i, j int) bool { return cand[i].W < cand[j].W })
	for _, s := range cand {
		if s.W >= targetW {
			return struct {
				URL  string
				W, H int
			}{s.URL, s.W, s.H}
		}
	}
	s := cand[len(cand)-1]
	return struct {
		URL  string
		W, H int
	}{s.URL, s.W, s.H}
}

//
// ===== Repo (DB) =====
//

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) UpsertPostWithMedia(ctx context.Context, p UnifiedPost, media []MediaItem) error {
	// per-post transaction timeout
	txCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := r.pool.BeginTx(txCtx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(txCtx)

	if _, err := upsertPost(txCtx, tx, p, media); err != nil {
		return err
	}

	return tx.Commit(txCtx)
}

func upsertPost(ctx context.Context, tx pgx.Tx, p UnifiedPost, media []MediaItem) (int64, error) {
	var views any
	if p.ViewsNullable == nil {
		views = nil
	} else {
		views = *p.ViewsNullable
	}
	var mediaJSON any
	if len(media) == 0 {
		mediaJSON = nil
	} else {
		b, err := json.Marshal(media)
		if err != nil {
			return 0, fmt.Errorf("marshal media: %w", err)
		}
		mediaJSON = b
	}

	var id int64
	err := tx.QueryRow(ctx, `
insert into posts (
  source, source_post_id,
  published_at, text, media, likes_count, views_count, comments_count,
  updated_at
) values (
  $1,$2,
  $3,$4,$5,$6,$7,$8,
  now()
)
on conflict (source, source_post_id) do update set
  published_at = excluded.published_at,
  text = excluded.text,
  media = excluded.media,
  likes_count = excluded.likes_count,
  views_count = excluded.views_count,
  comments_count = excluded.comments_count,
  updated_at = now()
returning id
`,
		p.Source, p.SourcePostID,
		p.PublishedAt, nullIfEmpty(p.Text), mediaJSON, p.Likes, views, p.CommentsCount,
	).Scan(&id)

	return id, err
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func intOrNull(v int) any {
	if v <= 0 {
		return nil
	}
	return v
}
