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
	return fmt.Sprintf("leaderboard member %s not found in leaderboard %s", mnfe.member, mnfe.leaderboard)
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
