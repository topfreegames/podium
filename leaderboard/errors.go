package leaderboard

import "fmt"

//MemberNotFoundError indicates member was not found in Redis
type MemberNotFoundError struct {
	LeaderboardID string
	MemberID      string
}

func (e *MemberNotFoundError) Error() string {
	return fmt.Sprintf("Could not find data for member %s in leaderboard %s.", e.MemberID, e.LeaderboardID)
}

//NewMemberNotFound returns a new error for member not found
func NewMemberNotFound(leaderboardID, memberID string) *MemberNotFoundError {
	return &MemberNotFoundError{
		LeaderboardID: leaderboardID,
		MemberID:      memberID,
	}
}
