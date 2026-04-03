package feed

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/auth"
	"github.com/goritskimihail/mudro/internal/chat"
	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/posts"
	mhttputil "github.com/goritskimihail/mudro/pkg/httputil"
)

// Server is the HTTP delivery layer for the feed domain.
type Server struct {
	postsSvc         *posts.Service
	authSvc          *auth.Service
	chatHandler      *chat.Handler
	tgVisiblePostIDs []string
	httpClient       *http.Client
	casinoServiceURL string
}

// NewServer constructs a Server with the provided service dependencies.
func NewServer(postsSvc *posts.Service, chatHandler *chat.Handler, authSvc *auth.Service) *Server {
	return &Server{
		postsSvc:         postsSvc,
		authSvc:          authSvc,
		chatHandler:      chatHandler,
		httpClient:       &http.Client{Timeout: 10 * time.Second},
		casinoServiceURL: config.CasinoServiceURL(),
	}
}

// Router registers all routes and returns the wrapped HTTP handler.
func (s *Server) Router() http.Handler {
	mediaRoot := strings.TrimSpace(os.Getenv("MEDIA_ROOT"))
	if mediaRoot == "" {
		mediaRoot = filepath.Join(config.RepoRoot(), "data", "nu")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/feed", http.StatusTemporaryRedirect)
	})
	mux.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(mediaRoot))))
	mux.HandleFunc("/healthz", mhttputil.HandleHealth("feed-api"))
	mux.HandleFunc("/api/posts", s.handlePosts)
	mux.HandleFunc("/api/front", s.handleFront)
	if s.authSvc != nil {
		mux.HandleFunc("/api/auth/login", s.handleAuthLogin)
		mux.HandleFunc("/api/auth/register", s.handleAuthRegister)
		mux.HandleFunc("/api/auth/me", s.handleAuthMe)
		mux.HandleFunc("/api/casino/balance", s.handleCasinoBalance)
		mux.HandleFunc("/api/casino/history", s.handleCasinoHistory)
		mux.HandleFunc("/api/casino/spin", s.handleCasinoSpin)
		mux.HandleFunc("/api/casino/config", s.handleCasinoConfig)
	}
	if s.chatHandler != nil {
		mux.HandleFunc("/api/chat/ws", s.chatHandler.HandleWS)
		mux.HandleFunc("/api/chat/history", s.chatHandler.HandleHistory)
	}
	mux.HandleFunc("/feed", s.handleFeed)

	return mhttputil.CORS(mhttputil.CORSConfig{SecurityHeaders: true})(mux)
}

// --- Wrapper delegation methods ---

// loadPosts delegates to postsSvc.LoadPosts; used by integration tests and handleFeed.
func (s *Server) loadPosts(ctx context.Context, beforeTS *time.Time, beforeID *int64, page *int, limit int, source string, sortOrder string) ([]posts.Post, *posts.Cursor, error) {
	if s.postsSvc == nil {
		return nil, nil, nil
	}
	return s.postsSvc.LoadPosts(ctx, beforeTS, beforeID, page, limit, source, posts.SortOrder(sortOrder), "")
}

func (s *Server) loadSourceStats(ctx context.Context) ([]posts.SourceStat, error) {
	if s.postsSvc == nil {
		return nil, nil
	}
	return s.postsSvc.LoadSourceStats(ctx)
}

func (s *Server) countVisiblePosts(ctx context.Context, total *int64) error {
	if s.postsSvc == nil {
		if total != nil {
			*total = 0
		}
		return nil
	}
	count, err := s.postsSvc.CountVisiblePosts(ctx)
	if err != nil {
		return err
	}
	if total != nil {
		*total = count
	}
	return nil
}

// buildPostsVisibilityWhere builds a SQL WHERE clause.
// Kept on Server because it is tested directly in server_test.go.
func (s *Server) buildPostsVisibilityWhere(source string, query *string) (string, []any) {
	conditions := make([]string, 0, 3)
	args := []any{}
	if source != "" {
		args = append(args, source)
		conditions = append(conditions, fmt.Sprintf("source = $%d", len(args)))
	}
	if query != nil {
		q := strings.TrimSpace(*query)
		if q != "" {
			args = append(args, "%"+strings.ToLower(q)+"%")
			conditions = append(conditions, fmt.Sprintf("LOWER(text) LIKE $%d", len(args)))
		}
	}
	if len(s.tgVisiblePostIDs) > 0 && (source == "" || source == "tg") {
		args = append(args, s.tgVisiblePostIDs)
		switch source {
		case "tg":
			conditions = append(conditions, fmt.Sprintf("source_post_id = any($%d)", len(args)))
		case "":
			conditions = append(conditions, fmt.Sprintf("(source <> 'tg' or source_post_id = any($%d))", len(args)))
		}
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " where " + strings.Join(conditions, " and "), args
}
