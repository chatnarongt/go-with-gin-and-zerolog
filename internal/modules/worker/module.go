package worker

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/logger"
)

type Job = internal.Job

type registry struct {
	mu   sync.RWMutex
	jobs map[string]Job
}

func newRegistry() *registry {
	return &registry{jobs: make(map[string]Job)}
}

func (r *registry) RegisterJob(name string, handler Job) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[name] = handler
}

func (r *registry) Get(name string) (Job, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	j, ok := r.jobs[name]
	return j, ok
}

type ModuleOptions struct {
	Logger  logger.ModuleOptions
	Config  config.ModuleOptions
	Imports []internal.Module
	Jobs    map[string]Job
	Args    []string
}

type Module struct {
	config        *config.Module
	logger        *logger.Module
	loggerOptions logger.ModuleOptions
	imports       []internal.Module
	registry      *registry
	overrideJobs  map[string]Job
	args          []string
}

var _ internal.Module = (*Module)(nil)

func NewModule(options ModuleOptions) *Module {
	overrideJobs := make(map[string]Job, len(options.Jobs))
	for name, handler := range options.Jobs {
		overrideJobs[name] = handler
	}
	return &Module{
		config:        config.NewModule(options.Config),
		logger:        logger.NewModule(options.Logger),
		loggerOptions: options.Logger,
		imports:       options.Imports,
		registry:      newRegistry(),
		overrideJobs:  overrideJobs,
		args:          options.Args,
	}
}

func (m *Module) coreModules() []internal.Module {
	return []internal.Module{m.logger, m.config}
}

func (m *Module) allModules() []internal.Module {
	return append(m.coreModules(), m.imports...)
}

func (m *Module) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return m.StartContext(ctx)
}

type jobLogHook struct {
	jobID   string
	jobName string
}

func (h jobLogHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	if h.jobID != "" {
		e.Str("jobId", h.jobID)
	}
	if h.jobName != "" {
		e.Str("job", h.jobName)
	}
}

func (m *Module) StartContext(ctx context.Context) error {
	injector := do.New()
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	jobName, err := m.parseJobFlag()
	if err != nil {
		log.Error().Err(err).Msg("Parse job argument")
		return err
	}
	if jobName == "" {
		err := errors.New("missing required --job argument")
		log.Error().Err(err).Msg("Validate job argument")
		return err
	}

	jobID, err := internal.NewID()
	if err != nil {
		log.Error().Err(err).Msg("Generate job ID")
		return err
	}

	m.logger = logger.NewModule(logger.ModuleOptions{
		Hooks: append([]zerolog.Hook{jobLogHook{jobID: jobID, jobName: jobName}}, m.loggerOptions.Hooks...),
	})

	if err := m.registerCore(injector); err != nil {
		log.Error().Err(err).Msg("Register worker modules")
		return err
	}

	_, err = do.Invoke[*config.Config](injector)
	if err != nil {
		log.Error().Err(err).Msg("Load worker config")
		return err
	}
	log = *do.MustInvoke[*zerolog.Logger](injector)

	if err := m.registerImports(injector); err != nil {
		log.Error().Err(err).Msg("Register worker modules")
		return err
	}
	for name, handler := range m.overrideJobs {
		m.registry.RegisterJob(name, handler)
	}
	if err := m.initializeModules(); err != nil {
		log.Error().Err(err).Msg("Initialize worker modules")
		return err
	}

	job, ok := m.registry.Get(jobName)
	if !ok {
		err := fmt.Errorf("unknown job %q", jobName)
		log.Error().Err(err).Msg("Resolve worker job")
		return err
	}

	log.Info().Msg("Starting worker job")
	jobErr := job(ctx, injector)
	if jobErr != nil {
		log.Error().Err(jobErr).Msg("Worker job failed")
	} else {
		log.Info().Msg("Worker job completed successfully")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.shutdown(shutdownContext, injector); err != nil {
		log.Error().Err(err).Msg("Worker shutdown failed")
		return errors.Join(jobErr, err)
	}
	log.Info().Msg("Worker shut down successfully")
	return jobErr
}

func (m *Module) parseJobFlag() (string, error) {
	args := m.args
	if args == nil {
		args = os.Args[1:]
	}

	var jobName string
	flags := flag.NewFlagSet("worker", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&jobName, "job", "", "Job name to execute")

	if err := flags.Parse(args); err != nil {
		return "", err
	}
	return strings.TrimSpace(jobName), nil
}

func (m *Module) registerCore(i do.Injector) error {
	do.ProvideValue[internal.JobRegistry](i, m.registry)
	for _, module := range m.coreModules() {
		if err := module.Register(i, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) registerImports(i do.Injector) error {
	for _, module := range m.imports {
		if err := module.Register(i, nil); err != nil {
			return err
		}
		if registrar, ok := module.(internal.JobRegistrar); ok {
			registrar.RegisterJobs(m.registry)
		}
	}
	return nil
}

func (m *Module) initializeModules() error {
	for _, module := range m.allModules() {
		if lifecycle, ok := module.(internal.OnModuleInit); ok {
			if err := lifecycle.OnModuleInit(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Module) Register(i do.Injector, _ *gin.Engine) error {
	if err := m.registerCore(i); err != nil {
		return err
	}
	if err := m.registerImports(i); err != nil {
		return err
	}
	return m.initializeModules()
}

func (m *Module) destroy(ctx context.Context) error {
	var destroyErrors []error

	for index := len(m.imports) - 1; index >= 0; index-- {
		lifecycle, ok := m.imports[index].(internal.OnModuleDestroy)
		if !ok {
			continue
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- lifecycle.OnModuleDestroy(ctx)
		}()

		select {
		case err := <-errCh:
			if err != nil {
				destroyErrors = append(destroyErrors, err)
			}
		case <-ctx.Done():
			return errors.Join(append(destroyErrors, ctx.Err())...)
		}
	}

	return errors.Join(destroyErrors...)
}

func (m *Module) shutdown(ctx context.Context, injector do.Injector) error {
	injectorReport := injector.ShutdownWithContext(ctx)
	var injectorErr error
	if len(injectorReport.Errors) > 0 {
		injectorErr = injectorReport
	}
	moduleErr := m.destroy(ctx)
	return errors.Join(injectorErr, moduleErr)
}
