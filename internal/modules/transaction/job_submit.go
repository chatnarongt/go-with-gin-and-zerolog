package transaction

import (
	"context"

	"github.com/samber/do/v2"
)

func (j *Job) runSubmit(ctx context.Context, i do.Injector) error {
	res := j.probe.GetReadiness(ctx)
	j.logger.Trace().Msg("Running transaction submit job: " + string(res.DatabaseMain))
	return nil
}
