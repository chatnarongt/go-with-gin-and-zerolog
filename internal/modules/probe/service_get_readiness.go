package probe

import (
	"context"
	"time"
)

const readinessPingTimeout = time.Second

type ReadinessStatus string

const (
	ReadinessStatusOK       ReadinessStatus = "OK"
	ReadinessStatusNotReady ReadinessStatus = "NOT_READY"
)

type GetReadinessResponseBody struct {
	Database ReadinessStatus `json:"database"`
	Service1 ReadinessStatus `json:"service1"`
	Service2 ReadinessStatus `json:"service2"`
}

func (s *Service) GetReadiness(ctx context.Context) GetReadinessResponseBody {
	pingContext, cancel := context.WithTimeout(ctx, readinessPingTimeout)
	defer cancel()

	databaseStatus := ReadinessStatusOK
	if err := s.db.PingContext(pingContext); err != nil {
		databaseStatus = ReadinessStatusNotReady
	}

	return GetReadinessResponseBody{
		Database: databaseStatus,
		Service1: ReadinessStatusOK,
		Service2: ReadinessStatusOK,
	}
}
