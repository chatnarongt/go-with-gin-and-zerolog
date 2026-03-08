package schedule

import (
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Module struct {
	db *database.Module
	c  *cron.Cron
}

func NewModule(db *database.Module) *Module {
	m := &Module{
		db: db,
		c:  cron.New(),
	}

	m.registerJobs()

	m.c.Start()

	log.Debug().Msg("Schedule Module initialized successfully")

	return m
}

func (m *Module) registerJobs() {
	m.CronCheckDbAlive()
}

func (m *Module) Cleanup() {
	log.Debug().Msg("Shutting down Schedule module...")
	ctx := m.c.Stop()
	<-ctx.Done() // wait for any running job to finish
	log.Debug().Msg("Schedule module shutdown complete")
}
