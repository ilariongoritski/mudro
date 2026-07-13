package authapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultServiceCatalog = "feed-api=http://api:8080/healthz,auth-api=http://auth-api:8087/healthz,bff-web=http://bff-web:8086/healthz,api-gateway=http://api-gateway:8085/healthz,orchestration-api=http://orchestration-api:8088/healthz,movie-catalog=http://movie-catalog:8091/healthz,casino-api=http://casino-api:8081/healthz,rag-api=http://rag-api:8092/healthz"

type runtimeDashboard struct {
	Providers []runtimeProvider `json:"providers"`
	Limits    runtimeLimits     `json:"limits"`
	Services  []runtimeService  `json:"services"`
}

type runtimeProvider struct {
	Name       string `json:"name"`
	Configured bool   `json:"configured"`
	Model      string `json:"model,omitempty"`
	Limit      string `json:"limit,omitempty"`
}

type runtimeLimits struct {
	RequestsPerSecond string `json:"requests_per_second,omitempty"`
	Burst             string `json:"burst,omitempty"`
}

type runtimeService struct {
	Name      string `json:"name"`
	HealthURL string `json:"-"`
	Status    string `json:"status"`
}

func (h *AdminHandlers) HandleGetRuntime(w http.ResponseWriter, r *http.Request) {
	dashboard := buildRuntimeDashboard(os.Getenv)
	dashboard.Services = h.checkServices(r.Context(), dashboard.Services)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dashboard)
}

func buildRuntimeDashboard(getenv func(string) string) runtimeDashboard {
	return runtimeDashboard{
		Providers: []runtimeProvider{
			{
				Name:       valueOr(getenv("DASHBOARD_LLM_LABEL"), "LLM"),
				Configured: parseBool(getenv("DASHBOARD_LLM_CONFIGURED")),
				Model:      strings.TrimSpace(getenv("DASHBOARD_LLM_MODEL")),
				Limit:      strings.TrimSpace(getenv("DASHBOARD_LLM_LIMIT")),
			},
			{
				Name:       valueOr(getenv("DASHBOARD_RAG_LABEL"), "Documentation RAG"),
				Configured: parseBool(getenv("DASHBOARD_RAG_CONFIGURED")),
				Model:      strings.TrimSpace(getenv("DASHBOARD_RAG_MODEL")),
				Limit:      strings.TrimSpace(getenv("DASHBOARD_RAG_LIMIT")),
			},
		},
		Limits: runtimeLimits{
			RequestsPerSecond: strings.TrimSpace(getenv("API_RATE_LIMIT_RPS")),
			Burst:             strings.TrimSpace(getenv("API_RATE_LIMIT_BURST")),
		},
		Services: parseServiceCatalog(valueOr(getenv("DASHBOARD_SERVICE_CATALOG"), defaultServiceCatalog)),
	}
}

func (h *AdminHandlers) checkServices(ctx context.Context, services []runtimeService) []runtimeService {
	client := &http.Client{Timeout: 2 * time.Second}
	for index := range services {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, services[index].HealthURL, nil)
		if err != nil {
			services[index].Status = "unavailable"
			continue
		}
		response, err := client.Do(request)
		if err != nil {
			services[index].Status = "unavailable"
			continue
		}
		response.Body.Close()
		if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
			services[index].Status = "healthy"
		} else {
			services[index].Status = "unavailable"
		}
		services[index].HealthURL = ""
	}
	return services
}

func parseServiceCatalog(value string) []runtimeService {
	entries := strings.Split(value, ",")
	services := make([]runtimeService, 0, len(entries))
	for _, entry := range entries {
		name, healthURL, found := strings.Cut(strings.TrimSpace(entry), "=")
		name = strings.TrimSpace(name)
		healthURL = strings.TrimSpace(healthURL)
		if !found || name == "" || healthURL == "" {
			continue
		}
		services = append(services, runtimeService{Name: name, HealthURL: healthURL, Status: "unknown"})
	}
	return services
}

func valueOr(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
