package probe_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/probe"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
)

func TestProbeReadiness_CamelCaseResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer os.Remove("./data/test_probe_main.db")
	defer os.Remove("./data/test_probe_analytics.db")

	injector := do.New()
	nopLogger := zerolog.Nop()
	do.ProvideValue(injector, &nopLogger)

	cfg := &config.Config{
		Databases: config.DatabasesConfig{
			Main: config.DatabaseConnectionConfig{
				Driver: "sqlite",
				DSN:    "file:./data/test_probe_main.db",
			},
			Analytics: config.DatabaseConnectionConfig{
				Driver: "sqlite",
				DSN:    "file:./data/test_probe_analytics.db",
			},
		},
	}
	do.ProvideValue(injector, cfg)

	dbModule := database.NewModule()
	if err := dbModule.Register(injector, nil); err != nil {
		t.Fatalf("register db module: %v", err)
	}

	router := gin.New()
	probeModule := probe.NewModule()
	if err := probeModule.Register(injector, router); err != nil {
		t.Fatalf("register probe module: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/probe/readiness", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if res["databaseMain"] != "OK" {
		t.Errorf("expected databaseMain=OK, got %v", res["databaseMain"])
	}
	if res["databaseAnalytics"] != "OK" {
		t.Errorf("expected databaseAnalytics=OK, got %v", res["databaseAnalytics"])
	}
}
