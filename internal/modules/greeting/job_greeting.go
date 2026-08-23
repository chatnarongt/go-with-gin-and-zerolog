package greeting

import (
	"context"

	"github.com/samber/do/v2"
)

func (j *Job) runGreeting(ctx context.Context, i do.Injector) error {
	getDatabaseTimeQuery := `SELECT datetime()`

	var currentTime string
	err := j.db.Main.QueryRowContext(ctx, getDatabaseTimeQuery).Scan(&currentTime)
	if err != nil {
		j.logger.Error().Err(err).Msg("Failed to get current time from database")
		return err
	}

	j.logger.Info().Msg("Hello! The current time from the database is: " + currentTime)
	return nil
}
