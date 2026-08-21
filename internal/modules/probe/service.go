package probe

import (
	"database/sql"

	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
)

type Service struct {
	log *zerolog.Logger
	db  *sql.DB
}

func NewService(i do.Injector) (*Service, error) {
	return &Service{
		log: do.MustInvoke[*zerolog.Logger](i),
		db:  do.MustInvoke[*sql.DB](i),
	}, nil
}

func (s *Service) OnModuleInit() error {
	return nil
}
