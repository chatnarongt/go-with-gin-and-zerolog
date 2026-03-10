package schedule

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

func (m *Module) CronCheckDbAlive() {
	if _, err := m.c.AddFunc("0 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := m.db.DB.PingContext(ctx); err != nil {
			log.Error().Err(err).Msg("CronCheckDbAlive: database ping failed")
		} else {
			log.Info().Msg("CronCheckDbAlive: database ping successful")
		}
	}); err != nil {
		log.Fatal().Err(err).Msg("Failed to register CronCheckDbAlive")
	}
}
