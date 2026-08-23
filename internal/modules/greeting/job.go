package greeting

import (
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
)

type Job struct {
	logger *zerolog.Logger
	db     *database.Databases
}

var _ internal.JobRegistrar = (*Job)(nil)

func NewJob(i do.Injector) (*Job, error) {
	return &Job{
		logger: do.MustInvoke[*zerolog.Logger](i),
		db:     do.MustInvoke[*database.Databases](i),
	}, nil
}

func (j *Job) RegisterJobs(registry internal.JobRegistry) {
	registry.RegisterJob("greeting:greeting", j.runGreeting)
}
