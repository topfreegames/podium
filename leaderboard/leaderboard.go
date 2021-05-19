package leaderboard

import (
	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/service"
)

var _ service.Leaderboard = &service.Service{}
var _ database.Database = &database.Redis{}
var _ database.Expiration = &database.Redis{}
