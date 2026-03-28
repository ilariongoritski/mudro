package wiring

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/posts"
	"github.com/goritskimihail/mudro/internal/tgexport"
	"github.com/goritskimihail/mudro/services/feed-api/internal/feed"
)

// NewHandler builds the legacy HTTP stack for feed-api and preserves current runtime behavior.
func NewHandler(pool *pgxpool.Pool) (http.Handler, error) {

	var tgVisiblePostIDs []string
	if ids, path, err := tgexport.LoadVisibleSourcePostIDsFromRepo(config.RepoRoot()); err == nil && len(ids) > 0 {
		tgVisiblePostIDs = ids
		log.Printf("main: loaded telegram visibility filter (%d ids) from %s", len(ids), path)
	} else if err != nil {
		log.Printf("main: telegram visibility filter disabled: %v", err)
	}

	postsSvc := posts.NewService(pool, tgVisiblePostIDs)
	return feed.NewServer(pool, postsSvc).Router(), nil
}
