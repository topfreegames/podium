package service

import (
	"github.com/topfreegames/podium/leaderboard/v2/database"
)

// Service holds all dependencies to leaderboard execute your operations
type Service struct {
	database.Database
}

// NewService instantiate a new Service
func NewService(database database.Database) *Service {
	return &Service{database}
}
