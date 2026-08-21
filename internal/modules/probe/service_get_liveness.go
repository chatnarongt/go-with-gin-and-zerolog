package probe

type GetLivenessResponseBody string

const (
	GetLivenessResponseBodyOK GetLivenessResponseBody = "OK"
)

func (s *Service) GetLiveness() GetLivenessResponseBody {
	return GetLivenessResponseBodyOK
}
