package transaction

import (
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/probe"
)

type Job struct {
	logger *zerolog.Logger
	db     *database.Databases
	probe  *probe.Service
}

var _ internal.JobRegistrar = (*Job)(nil)

func NewJob(i do.Injector) (*Job, error) {
	return &Job{
		logger: do.MustInvoke[*zerolog.Logger](i),
		db:     do.MustInvoke[*database.Databases](i),
		probe:  do.MustInvoke[*probe.Service](i),
	}, nil
}

func (j *Job) RegisterJobs(registry internal.JobRegistry) {
	registry.RegisterJob("transaction:submit", j.runSubmit)
}
