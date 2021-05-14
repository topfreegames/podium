package database

import "fmt"

// GeneralError create a redis error that is not handled
type GeneralError struct {
	msg string
}

func (ue *GeneralError) Error() string {
	return fmt.Sprintf("database error: %s", ue.msg)
}

// NewGeneralError create a new redis error that isnt handled
func NewGeneralError(msg string) *GeneralError {
	return &GeneralError{msg: msg}
}

// InvalidOrderError is an error when an invalid order was gave
type InvalidOrderError struct {
	order string
}

func (ioe *InvalidOrderError) Error() string {
	return fmt.Sprintf("invalid order: %s", ioe.order)
}

// NewInvalidOrderError create a new InvalidOrderError
func NewInvalidOrderError(order string) *InvalidOrderError {
	return &InvalidOrderError{
		order: order,
	}
}

// MemberNotFoundError is an error throw when leaderboard not have member
type MemberNotFoundError struct {
	leaderboard string
	member      string
}

func (mnfe *MemberNotFoundError) Error() string {
	return fmt.Sprintf("member %s not found in leaderboard %s", mnfe.member, mnfe.leaderboard)
}

// NewMemberNotFoundError create a new MemberNotFoundError
func NewMemberNotFoundError(leaderboard, member string) *MemberNotFoundError {
	return &MemberNotFoundError{
		leaderboard: leaderboard,
		member:      member,
	}
}

// TTLNotFoundError is an error throw when key has not TTL
type TTLNotFoundError struct {
	leaderboard string
}

// NewTTLNotFoundError create a new KeyNotFoundError
func NewTTLNotFoundError(leaderboard string) *TTLNotFoundError {
	return &TTLNotFoundError{
		leaderboard: leaderboard,
	}
}

func (tnfe *TTLNotFoundError) Error() string {
	return fmt.Sprintf("ttl to leaderboard %s not found", tnfe.leaderboard)
}

// LeaderboardWithoutMemberToExpireError is an error throw when leaderboard doesn't have member to expire
type LeaderboardWithoutMemberToExpireError struct {
	leaderboard string
}

// NewLeaderboardWithoutMemberToExpireError create a new KeyNotFoundError
func NewLeaderboardWithoutMemberToExpireError(leaderboard string) *LeaderboardWithoutMemberToExpireError {
	return &LeaderboardWithoutMemberToExpireError{
		leaderboard: leaderboard,
	}
}

func (lwmtee *LeaderboardWithoutMemberToExpireError) Error() string {
	return fmt.Sprintf("leaderboard %s without member to expire", lwmtee.leaderboard)
}
