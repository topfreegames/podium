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
	return fmt.Sprintf("redis key %s not found", knfe.key)
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
	return fmt.Sprintf("redis key %s not found", knfe.key)
}

// MemberNotFoundError is an error throw when key has not Member
type MemberNotFoundError struct {
	key string
}

// NewMemberNotFoundError create a new KeyNotFoundError
func NewMemberNotFoundError(key string) *MemberNotFoundError {
	return &MemberNotFoundError{
		key: key,
	}
}

func (mnfe *MemberNotFoundError) Error() string {
	return fmt.Sprintf("redis key %s not found", mnfe.key)
}
