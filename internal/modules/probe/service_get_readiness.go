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
	DatabaseMain      ReadinessStatus `json:"databaseMain"`
	DatabaseAnalytics ReadinessStatus `json:"databaseAnalytics"`
	DatabaseLogging   ReadinessStatus `json:"databaseLogging"`
	Service1          ReadinessStatus `json:"service1"`
	Service2          ReadinessStatus `json:"service2"`
}

func (s *Service) GetReadiness(ctx context.Context) GetReadinessResponseBody {
	pingContext, cancel := context.WithTimeout(ctx, readinessPingTimeout)
	defer cancel()

	mainStatus := ReadinessStatusOK
	if err := s.db.Main.PingContext(pingContext); err != nil {
		mainStatus = ReadinessStatusNotReady
	}

	analyticsStatus := ReadinessStatusOK
	if err := s.db.Analytics.PingContext(pingContext); err != nil {
		analyticsStatus = ReadinessStatusNotReady
	}

	loggingStatus := ReadinessStatusNotReady
	if s.db.Logging != nil {
		if err := s.db.Logging.Client().Ping(pingContext, nil); err == nil {
			loggingStatus = ReadinessStatusOK
		}
	}

	return GetReadinessResponseBody{
		DatabaseMain:      mainStatus,
		DatabaseAnalytics: analyticsStatus,
		DatabaseLogging:   loggingStatus,
		Service1:          ReadinessStatusOK,
		Service2:          ReadinessStatusOK,
	}
}
