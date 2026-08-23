package greeting_test

import (
	"context"
	"testing"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/greeting"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/worker"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
)

func TestGreetingModule_RegisterAndExecuteJobs(t *testing.T) {
	t.Setenv("DB_LOGGING_DSN", "")

	app := worker.NewModule(worker.ModuleOptions{
		Args: []string{"--job=greeting:greeting"},
		Imports: []internal.Module{
			database.NewModule(),
			greeting.NewModule(),
		},
	})

	if err := app.StartContext(context.Background()); err != nil {
		t.Fatalf("expected greeting:greeting job to succeed, got %v", err)
	}
}

func TestGreetingJob_RegisterJobsDirectly(t *testing.T) {
	injector := do.New()
	nopLogger := zerolog.Nop()
	do.ProvideValue(injector, &nopLogger)
	do.ProvideValue(injector, &database.Databases{})

	job, err := greeting.NewJob(injector)
	if err != nil {
		t.Fatalf("new job failed: %v", err)
	}

	jobs := make(map[string]internal.Job)
	reg := &mockRegistry{jobs: jobs}
	job.RegisterJobs(reg)

	if _, ok := jobs["greeting:greeting"]; !ok {
		t.Error("expected greeting:greeting job to be registered")
	}
}

type mockRegistry struct {
	jobs map[string]internal.Job
}

func (m *mockRegistry) RegisterJob(name string, handler internal.Job) {
	m.jobs[name] = handler
}
