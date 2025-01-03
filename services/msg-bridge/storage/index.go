package storage

import (
	"fmt"
	GenericStorage "peergrine/utils/generic-storage"
)

const (
	REDIS_PREFIX_LINKCODE       = "message-linkcode:"
	REDIS_PREFIX_CLIENT_CHANNEL = "message-client-channel:"
)

type ClientSession struct {
	LinkCode     string
	ClientId     string
	ChannelId    string
	SessionBytes []byte
	ExpiresAt    int64
}

func (m ClientSession) GetKey() string {
	return m.LinkCode
}

func (m ClientSession) GetExpiresAt() int64 {
	return m.ExpiresAt
}

type Storage struct {
	*GenericStorage.Storage[ClientSession]
}

func New(channelId string, redisAddr string) (*Storage, error) {
	s, err := GenericStorage.New[ClientSession](channelId, redisAddr)

	if err != nil {
		return nil, err
	}
	storage := &Storage{s}
	return storage, nil
}

func (s *Storage) SetClientSession(session ClientSession) error {

	s.Local.Set(session)

	if s.Redis != nil {
		key := REDIS_PREFIX_LINKCODE + session.GetKey()
		return s.SetToRedis(key, session)
	}

	return nil
}

func (s *Storage) GetClientSession(linkCode string) (*ClientSession, error) {

	localSession := s.Local.Get(linkCode)

	if localSession != nil {
		return localSession, nil
	}

	if s.Redis != nil {
		key := REDIS_PREFIX_LINKCODE + linkCode
		return s.GetFromRedis(key)
	}

	return nil, fmt.Errorf("client session not found for link code: %s", linkCode)
}

func (s *Storage) ClientSessionExists(linkCode string) (bool, error) {

	localExists := s.Local.Exists(linkCode)

	if localExists {
		return localExists, nil
	}

	if s.Redis != nil {
		key := REDIS_PREFIX_LINKCODE + linkCode
		return s.Redis.Exists(key)
	}

	return false, nil
}

func (s *Storage) RemoveClientSession(linkCode string) error {

	s.Local.Remove(linkCode)

	if s.Redis != nil {
		key := REDIS_PREFIX_LINKCODE + linkCode
		return s.Redis.Del(key)
	}

	return nil
}

func (s *Storage) SetClientChannel(clientId string, channelId string) error {

	if s.Redis != nil {

		key := REDIS_PREFIX_CLIENT_CHANNEL + clientId

		content := []byte(channelId)
		err := s.Redis.Set(key, content, 0)
		if err != nil {
			return err
		}

	}

	return nil
}

func (s *Storage) GetClientChannel(clientId string) (string, error) {

	if s.Redis != nil {

		key := REDIS_PREFIX_CLIENT_CHANNEL + clientId

		content, err := s.Redis.Get(key)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve client channel for client ID: %s", clientId)
		}

		return string(content), nil
	}

	return "", fmt.Errorf("client channel not found for client ID: %s", clientId)
}

func (s Storage) RemoveClientChannel(clientId string) error {

	if s.Redis != nil {

		key := REDIS_PREFIX_CLIENT_CHANNEL + clientId
		return s.Redis.Del(key)

	}

	return nil
}
