package api

import (
	"encoding/json"
	"net/http"

	"github.com/goritskimihail/mudro/internal/config"
)

func (s *Server) handleWebfinger(w http.ResponseWriter, r *http.Request) {
	base := config.APIBaseURL()
	resp := map[string]any{
		"subject": "acct:mudro@" + r.Host,
		"links": []map[string]string{
			{
				"rel":  "self",
				"type": "application/activity+json",
				"href": base + "/api/activitypub/actor",
			},
		},
	}
	w.Header().Set("Content-Type", "application/jrd+json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleActor(w http.ResponseWriter, r *http.Request) {
	base := config.APIBaseURL()
	actor := map[string]any{
		"@context":          []string{"https://www.w3.org/ns/activitystreams"},
		"type":              "Service",
		"id":                base + "/api/activitypub/actor",
		"preferredUsername": "mudro",
		"name":              "Mudro Feed",
		"inbox":             base + "/api/activitypub/inbox",
		"outbox":            base + "/api/activitypub/outbox",
	}
	w.Header().Set("Content-Type", "application/activity+json")
	_ = json.NewEncoder(w).Encode(actor)
}
