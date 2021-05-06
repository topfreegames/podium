package redis

import "fmt"

// KeyNotFoundError is an error throw when key is not in redis
type KeyNotFoundError struct {
	key string
}

// NewKeyNotFoundError create a new KeyNotFoundError
func NewKeyNotFoundError(key string) *KeyNotFoundError {
	return &KeyNotFoundError{
		key: key,
	}
}

func (knfe *KeyNotFoundError) Error() string {
	return fmt.Sprintf("redis: key %s not found", knfe.key)
}

// TTLNotFoundError is an error throw when key has not TTL
type TTLNotFoundError struct {
	key string
}

// NewTTLNotFoundError create a new KeyNotFoundError
func NewTTLNotFoundError(key string) *TTLNotFoundError {
	return &TTLNotFoundError{
		key: key,
	}
}

func (knfe *TTLNotFoundError) Error() string {
	return fmt.Sprintf("redis: ttl to key %s not found", knfe.key)
}

// MemberNotFoundError is an error throw when key has not Member
type MemberNotFoundError struct {
	key    string
	member string
}

// NewMemberNotFoundError create a new KeyNotFoundError
func NewMemberNotFoundError(key, member string) *MemberNotFoundError {
	return &MemberNotFoundError{
		key: key,
	}
}

func (mnfe *MemberNotFoundError) Error() string {
	return fmt.Sprintf("redis: key %s not have member %s found", mnfe.key, mnfe.member)
}

// GeneralError create a redis error that is not handled
type GeneralError struct {
	msg string
}

func (ue *GeneralError) Error() string {
	return fmt.Sprintf("redis: error: %s", ue.msg)
}

// NewGeneralError create a new redis error that isnt handled
func NewGeneralError(msg string) *GeneralError {
	return &GeneralError{msg: msg}
}
