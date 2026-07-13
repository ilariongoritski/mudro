package authapi

import "testing"

func TestBuildRuntimeDashboardDoesNotExposeSecrets(t *testing.T) {
	values := map[string]string{
		"DASHBOARD_LLM_CONFIGURED": "true",
		"DASHBOARD_LLM_MODEL":      "example/chat-model",
		"DASHBOARD_LLM_LIMIT":      "100 requests/min",
		"DASHBOARD_RAG_CONFIGURED": "false",
		"DASHBOARD_RAG_MODEL":      "example/embedding-model",
		"DASHBOARD_RAG_LIMIT":      "",
		"DASHBOARD_SERVICE_CATALOG": "feed-api=http://feed-api:8080/healthz,auth-api=http://auth-api:8087/healthz",
		"LLM_API_KEY":              "must-not-be-returned",
	}

	dashboard := buildRuntimeDashboard(func(key string) string { return values[key] })

	if len(dashboard.Providers) != 2 {
		t.Fatalf("providers = %d, want 2", len(dashboard.Providers))
	}
	if dashboard.Providers[0].Name != "LLM" || !dashboard.Providers[0].Configured {
		t.Fatalf("provider = %+v", dashboard.Providers[0])
	}
	if dashboard.Providers[0].Model != "example/chat-model" || dashboard.Providers[0].Limit != "100 requests/min" {
		t.Fatalf("provider metadata = %+v", dashboard.Providers[0])
	}
	if len(dashboard.Services) != 2 || dashboard.Services[0].Name != "feed-api" {
		t.Fatalf("services = %+v", dashboard.Services)
	}
	if containsSecret(dashboard, values["LLM_API_KEY"]) {
		t.Fatal("dashboard must not contain API key values")
	}
}

func TestBuildRuntimeDashboardSkipsMalformedServiceEntries(t *testing.T) {
	dashboard := buildRuntimeDashboard(func(key string) string {
		if key == "DASHBOARD_SERVICE_CATALOG" {
			return "invalid, api = ,db=http://db:5432"
		}
		return ""
	})

	if len(dashboard.Services) != 1 || dashboard.Services[0].Name != "db" {
		t.Fatalf("services = %+v", dashboard.Services)
	}
}

func containsSecret(dashboard runtimeDashboard, secret string) bool {
	for _, provider := range dashboard.Providers {
		if provider.Name == secret || provider.Model == secret || provider.Limit == secret {
			return true
		}
	}
	for _, service := range dashboard.Services {
		if service.Name == secret || service.HealthURL == secret {
			return true
		}
	}
	return false
}
