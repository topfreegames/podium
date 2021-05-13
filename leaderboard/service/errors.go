package service

import "fmt"

// GeneralError is an error threw when a not handled error was found
type GeneralError struct {
	service string
	msg     string
}

func (ge *GeneralError) Error() string {
	return fmt.Sprintf("error on service %s: %s", ge.service, ge.msg)
}

// NewGeneralError create a new GeneralError
func NewGeneralError(service, msg string) *GeneralError {
	return &GeneralError{
		service: service,
		msg:     msg,
	}
}

// MemberNotFoundError is an error throw when leaderboard not have member
type MemberNotFoundError struct {
	leaderboard string
	member      string
}

// NewMemberNotFoundError create a new MemberNotFoundError
func NewMemberNotFoundError(leaderboard, member string) *MemberNotFoundError {
	return &MemberNotFoundError{
		leaderboard: leaderboard,
		member:      member,
	}
}

func (mnfe *MemberNotFoundError) Error() string {
	return fmt.Sprintf("Could not find data for member %s in leaderboard %s.", mnfe.member, mnfe.leaderboard)
}

// PageOutOfRangeError is an error threw when a page if out of permited range
type PageOutOfRangeError struct {
	page      int
	totalPage int
}

// NewPageOutOfRangeError create a new PageOutOfRangeError
func NewPageOutOfRangeError(page, totalPage int) *PageOutOfRangeError {
	return &PageOutOfRangeError{
		page:      page,
		totalPage: totalPage,
	}
}

func (poor *PageOutOfRangeError) Error() string {
	return fmt.Sprintf("page %d out of range (1, %d)", poor.page, poor.totalPage)
}

// LeaderboardExpiredError is an error threw when a not handled error was found
type LeaderboardExpiredError struct {
	leaderboard string
}

func (lee *LeaderboardExpiredError) Error() string {
	return fmt.Sprintf("Leaderboard expired error: %s", lee.leaderboard)
}

// NewLeaderboardExpiredError create a new LeaderboardExpiredError
func NewLeaderboardExpiredError(leaderboard string) *LeaderboardExpiredError {
	return &LeaderboardExpiredError{
		leaderboard: leaderboard,
	}
}

// PercentageError is an error threw when a not handled error was found
type PercentageError struct {
	percentage int
}

func (pe *PercentageError) Error() string {
	return fmt.Sprintf("percentage error: %d, it must be a valid integer between 1 and 100", pe.percentage)
}

// NewPercentageError create a new PercentageError
func NewPercentageError(percentage int) *PercentageError {
	return &PercentageError{
		percentage: percentage,
	}
}
