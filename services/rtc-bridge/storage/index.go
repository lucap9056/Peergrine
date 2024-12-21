package storage

import (
	"errors"
	GenericStorage "peergrine/utils/generic-storage"
)

const (
	REDIS_PREFIX_LINKCODE = "signal-linkcode:"
)

// Signal represents a communication signal with metadata such as LinkCode, ClientId, etc.
type Signal struct {
	LinkCode    string
	ClientId    string
	SignalBytes []byte
	ChannelId   int32
	ExpiresAt   int64
}

// NewSignal creates a new Signal instance with the provided client ID, signal data, and expiration time.
// Parameters:
//   - clientId (string): The client identifier for the signal.
//   - signal ([]byte): The signal data in byte format.
//   - expiresAt (int64): The timestamp when the signal expires.
//
// Returns:
//   - Signal: A new Signal object.
func NewSignal(clientId string, linkCode string, signal []byte, expiresAt int64) Signal {
	return Signal{
		ClientId:    clientId,
		LinkCode:    linkCode,
		SignalBytes: signal,
		ExpiresAt:   expiresAt,
	}
}

// SetLinkCode sets the LinkCode for the Signal.
// Parameters:
//   - linkCode (string): The link code to assign to the signal.
func (s *Signal) SetLinkCode(linkCode string) {
	s.LinkCode = linkCode
}

// SetChannelId sets the ChannelId for the Signal.
// Parameters:
//   - channelId (int32): The channel identifier to assign to the signal.
func (s *Signal) SetChannelId(channelId int32) {
	s.ChannelId = channelId
}

// GetKey returns the LinkCode as the key for the signal.
// Returns:
//   - string: The link code of the signal.
func (s Signal) GetKey() string {
	return s.LinkCode
}

// GetExpiresAt returns the expiration timestamp of the signal.
// Returns:
//   - int64: The expiration time of the signal.
func (s Signal) GetExpiresAt() int64 {
	return s.ExpiresAt
}

// Storage manages Signal storage and retrieval, using either Redis or local storage.
type Storage struct {
	*GenericStorage.Storage[Signal]
}

// New creates a new instance of the Storage to manage signals.
// Parameters:
//   - channelId (int32): The ID of the communication channel.
//   - redisAddr (string): The address of the Redis server (optional).
//
// Returns:
//   - *Storage: A new Storage instance or an error if Redis initialization fails.
func New(channelId int32, redisAddr string) (*Storage, error) {
	s, err := GenericStorage.New[Signal](channelId, redisAddr)
	if err != nil {
		return nil, err
	}
	storage := &Storage{s}
	return storage, nil
}

// SetSignal storages a Signal into either Redis or local storage.
// Parameters:
//   - linkCode (string): The link code of the signal.
//   - signal (Signal): The signal data to storage.
//
// Returns:
//   - error: nil if successful, otherwise an error message.
func (m *Storage) SetSignal(signal Signal) error {
	m.Local.Set(signal)

	if m.Redis != nil {
		key := REDIS_PREFIX_LINKCODE + signal.GetKey()
		return m.SetToRedis(key, signal)
	}

	return nil
}

// GetSignal retrieves a Signal from either Redis or local storage based on the provided link code.
// Parameters:
//   - linkCode (string): The link code of the signal to retrieve.
//
// Returns:
//   - *Signal: The retrieved signal, or nil if not found.
//   - error: nil if successful, otherwise an error message.
func (m *Storage) GetSignal(linkCode string) (*Signal, error) {

	localSignal := m.Local.Get(linkCode)

	if localSignal != nil {
		return localSignal, nil
	}

	if m.Redis != nil {
		key := REDIS_PREFIX_LINKCODE + linkCode
		return m.GetFromRedis(key)
	}

	return nil, errors.New("no storage manager configured")
}

// SignalExist checks if a signal with the given link code exists in either Redis or local storage.
// Parameters:
//   - linkCode (string): The link code to check.
//
// Returns:
//   - bool: true if the signal exists, false otherwise.
//   - error: nil if successful, otherwise an error message.
func (m *Storage) SignalExists(linkCode string) (bool, error) {

	localExists := m.Local.Exists(linkCode)

	if localExists {
		return localExists, nil
	}

	if m.Redis != nil {
		key := REDIS_PREFIX_LINKCODE + linkCode
		return m.Redis.Exists(key)
	}

	return false, nil
}

// RemoveSignal deletes a signal from either Redis or local storage.
// Parameters:
//   - linkCode (string): The link code of the signal to remove.
//
// Returns:
//   - error: nil if successful, otherwise an error message.
func (m *Storage) RemoveSignal(linkCode string) error {

	m.Local.Remove(linkCode)

	if m.Redis != nil {
		key := REDIS_PREFIX_LINKCODE + linkCode
		return m.Redis.Del(key)
	}

	return nil
}
