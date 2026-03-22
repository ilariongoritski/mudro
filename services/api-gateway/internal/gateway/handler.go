package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Config struct {
	FeedAPIURL string
	BFFWebURL  string
}

func NewHandler(cfg Config) (http.Handler, error) {
	feedURL, err := parseBaseURL(cfg.FeedAPIURL, "feed API")
	if err != nil {
		return nil, err
	}
	bffURL, err := parseBaseURL(cfg.BFFWebURL, "BFF web")
	if err != nil {
		return nil, err
	}

	feedProxy := newProxy(feedURL, func(path string) string {
		if strings.HasPrefix(path, "/api/v1/") {
			return "/api/" + strings.TrimPrefix(path, "/api/v1/")
		}
		return path
	})
	bffProxy := newProxy(bffURL, func(path string) string { return path })

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/api/v1/healthz", handleHealth)
	mux.Handle("/api/bff/web/v1/", bffProxy)
	mux.Handle("/api/v1/", feedProxy)
	return mux, nil
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func parseBaseURL(raw, label string) (*url.URL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, fmt.Errorf("%s URL is required", label)
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("%s URL parse: %w", label, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%s URL must include scheme and host", label)
	}
	return parsed, nil
}

func newProxy(target *url.URL, rewritePath func(string) string) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	defaultDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		defaultDirector(req)
		req.URL.Path = rewritePath(req.URL.Path)
		req.Host = target.Host
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		http.Error(w, "upstream unavailable: "+err.Error(), http.StatusBadGateway)
	}
	return proxy
}
