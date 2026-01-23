package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/99designs/keyring"
)

const (
	serviceName = "figma-cli"
	keyName     = "token"
)

// keyringStorage implements Storage using the system keyring.
type keyringStorage struct {
	ring keyring.Keyring
}

// NewKeyringStorage creates a new keyring-based token storage.
func NewKeyringStorage() (Storage, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: serviceName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}
	return &keyringStorage{ring: ring}, nil
}

// Store saves a token to the keyring.
func (s *keyringStorage) Store(ctx context.Context, token *Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	err = s.ring.Set(keyring.Item{
		Key:  keyName,
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("failed to store token in keyring: %w", err)
	}
	return nil
}

// Retrieve reads a token from the keyring.
func (s *keyringStorage) Retrieve(ctx context.Context) (*Token, error) {
	item, err := s.ring.Get(keyName)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve token from keyring: %w", err)
	}
	var token Token
	if err := json.Unmarshal(item.Data, &token); err != nil {
		// If unmarshal fails, maybe the stored format is different (legacy).
		// Try to treat as plain PAT token.
		token = Token{
			Type:         TokenTypePAT,
			AccessToken:  string(item.Data),
			RefreshToken: "",
			ExpiresAt:    nil,
		}
	}
	return &token, nil
}

// Clear removes the token from the keyring.
func (s *keyringStorage) Clear(ctx context.Context) error {
	err := s.ring.Remove(keyName)
	if err != nil && err != keyring.ErrKeyNotFound {
		return fmt.Errorf("failed to remove token from keyring: %w", err)
	}
	return nil
}
