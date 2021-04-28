package service

import "context"

const healthcheckServiceLabel = "healthcheck"

// Healthcheck check service dependencies if it is running
func (s *Service) Healthcheck(ctx context.Context) error {
	err := s.Database.Healthcheck(ctx)
	if err != nil {
		return NewGeneralError(healthcheckServiceLabel, err.Error())
	}

	return nil
}
