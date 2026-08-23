package worker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/swagger"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/worker"
	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type testModule struct {
	initCalled    bool
	destroyCalled bool
}

func (m *testModule) Register(i do.Injector, _ *gin.Engine) error {
	do.ProvideValue(i, "test-value")
	return nil
}

func (m *testModule) OnModuleInit() error {
	m.initCalled = true
	return nil
}

func (m *testModule) OnModuleDestroy(ctx context.Context) error {
	m.destroyCalled = true
	return nil
}

var (
	_ internal.Module          = (*testModule)(nil)
	_ internal.OnModuleInit    = (*testModule)(nil)
	_ internal.OnModuleDestroy = (*testModule)(nil)
)

type testJobModule struct {
	called bool
}

func (m *testJobModule) Register(i do.Injector, _ *gin.Engine) error {
	if registry, err := do.Invoke[internal.JobRegistry](i); err == nil && registry != nil {
		registry.RegisterJob("module:job", func(ctx context.Context, i do.Injector) error {
			m.called = true
			return nil
		})
	}
	return nil
}

type testRegistrarModule struct {
	called bool
}

func (m *testRegistrarModule) Register(i do.Injector, _ *gin.Engine) error {
	return nil
}

func (m *testRegistrarModule) RegisterJobs(registry internal.JobRegistry) {
	registry.RegisterJob("registrar:job", func(ctx context.Context, i do.Injector) error {
		m.called = true
		return nil
	})
}

var (
	_ internal.Module       = (*testRegistrarModule)(nil)
	_ internal.JobRegistrar = (*testRegistrarModule)(nil)
)

func TestWorkerModule_MissingJobFlag(t *testing.T) {
	w := worker.NewModule(worker.ModuleOptions{
		Args: []string{},
		Jobs: map[string]worker.Job{
			"Greeting": func(ctx context.Context, i do.Injector) error {
				return nil
			},
		},
	})

	err := w.StartContext(context.Background())
	if err == nil {
		t.Fatal("expected error for missing --job, got nil")
	}
	if err.Error() != "missing required --job argument" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestWorkerModule_UnknownJob(t *testing.T) {
	w := worker.NewModule(worker.ModuleOptions{
		Args: []string{"--job=NonExistent"},
		Jobs: map[string]worker.Job{
			"Greeting": func(ctx context.Context, i do.Injector) error {
				return nil
			},
		},
	})

	err := w.StartContext(context.Background())
	if err == nil {
		t.Fatal("expected error for unknown job, got nil")
	}
}

func TestWorkerModule_ExecuteJobWithImportsAndLifecycle(t *testing.T) {
	executed := false
	tm := &testModule{}

	w := worker.NewModule(worker.ModuleOptions{
		Args: []string{"--job=Greeting"},
		Imports: []internal.Module{
			tm,
		},
		Jobs: map[string]worker.Job{
			"Greeting": func(ctx context.Context, i do.Injector) error {
				val, err := do.Invoke[string](i)
				if err != nil {
					return err
				}
				if val != "test-value" {
					t.Errorf("expected test-value, got %s", val)
				}
				executed = true
				return nil
			},
		},
	})

	err := w.StartContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error running worker: %v", err)
	}

	if !executed {
		t.Error("expected job to be executed")
	}
	if !tm.initCalled {
		t.Error("expected module OnModuleInit to be called")
	}
	if !tm.destroyCalled {
		t.Error("expected module OnModuleDestroy to be called")
	}
}

func TestWorkerModule_AutoRegisteredJobViaDI(t *testing.T) {
	jm := &testJobModule{}
	w := worker.NewModule(worker.ModuleOptions{
		Args:    []string{"--job=module:job"},
		Imports: []internal.Module{jm},
	})

	err := w.StartContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !jm.called {
		t.Fatal("expected auto-registered module job to be executed")
	}
}

func TestWorkerModule_AutoRegisteredJobViaRegistrar(t *testing.T) {
	rm := &testRegistrarModule{}
	w := worker.NewModule(worker.ModuleOptions{
		Args:    []string{"--job=registrar:job"},
		Imports: []internal.Module{rm},
	})

	err := w.StartContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rm.called {
		t.Fatal("expected auto-registered registrar job to be executed")
	}
}

func TestWorkerModule_OptionJobsOverride(t *testing.T) {
	overrideCalled := false
	jm := &testJobModule{}
	w := worker.NewModule(worker.ModuleOptions{
		Args:    []string{"--job=module:job"},
		Imports: []internal.Module{jm},
		Jobs: map[string]worker.Job{
			"module:job": func(ctx context.Context, i do.Injector) error {
				overrideCalled = true
				return nil
			},
		},
	})

	err := w.StartContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !overrideCalled {
		t.Fatal("expected options.Jobs to override module job")
	}
	if jm.called {
		t.Fatal("expected module job to be overridden")
	}
}

func TestWorkerModule_NilRouterSafetyWithExistingModules(t *testing.T) {
	executed := false
	w := worker.NewModule(worker.ModuleOptions{
		Args: []string{"--job=test"},
		Imports: []internal.Module{
			swagger.NewModule(swagger.ModuleOptions{
				Title:   "test",
				Version: "1.0",
			}),
		},
		Jobs: map[string]worker.Job{
			"test": func(ctx context.Context, i do.Injector) error {
				executed = true
				return nil
			},
		},
	})

	err := w.StartContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error running worker with swagger module: %v", err)
	}
	if !executed {
		t.Fatal("expected test job to execute successfully")
	}
}

func TestWorkerModule_JobError(t *testing.T) {
	expectedErr := errors.New("job failed")
	w := worker.NewModule(worker.ModuleOptions{
		Args: []string{"--job=FailingJob"},
		Jobs: map[string]worker.Job{
			"FailingJob": func(ctx context.Context, i do.Injector) error {
				return expectedErr
			},
		},
	})

	err := w.StartContext(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected job error %v, got %v", expectedErr, err)
	}
}
