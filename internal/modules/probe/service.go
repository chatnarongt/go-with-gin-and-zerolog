package probe

import (
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
)

type Service struct {
	log *zerolog.Logger
	db  *database.Databases
}

func NewService(i do.Injector) (*Service, error) {
	return &Service{
		log: do.MustInvoke[*zerolog.Logger](i),
		db:  do.MustInvoke[*database.Databases](i),
	}, nil
}

func (s *Service) OnModuleInit() error {
	return nil
}
