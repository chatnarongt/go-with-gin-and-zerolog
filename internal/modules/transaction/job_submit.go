package transaction

import (
	"context"

	"github.com/samber/do/v2"
)

func (j *Job) runSubmit(ctx context.Context, i do.Injector) error {
	res := j.probe.GetLiveness()
	j.logger.Trace().Msg("Running transaction submit job: " + string(res))
	return nil
}
