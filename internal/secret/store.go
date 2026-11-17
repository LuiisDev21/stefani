// Package secret stores credentials in the operating system keychain. Nothing
// here ever writes a key to disk or to the database.
package secret

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

// ErrUnavailable means the platform keychain could not be reached. Callers must
// treat it as "the user has to configure the key again", never as "no key".
var ErrUnavailable = errors.New("system keychain unavailable")

// ErrNotFound means the key is simply not stored yet.
var ErrNotFound = errors.New("credential not found")

// Key names used by the application.
const (
	KeyLLMAPIKey = "llm-api-key"
)

// Store is the credential boundary. Tests use a fake implementation.
type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
	Available() bool
}

type keyringStore struct {
	service string
}

// New returns the OS-backed store, namespaced by service.
func New(service string) Store { return &keyringStore{service: service} }

func (s *keyringStore) Get(key string) (string, error) {
	value, err := keyring.Get(s.service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return value, nil
}

func (s *keyringStore) Set(key, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("refusing to store an empty credential")
	}
	if err := keyring.Set(s.service, key, value); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (s *keyringStore) Delete(key string) error {
	err := keyring.Delete(s.service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (s *keyringStore) Available() bool {
	// A successful read of a key that cannot exist proves the backend works;
	// ErrNotFound is the healthy answer.
	_, err := keyring.Get(s.service, "__probe__")
	return errors.Is(err, keyring.ErrNotFound)
}

// Memory is an in-memory Store for tests and for the browser E2E host, where no
// OS keychain exists.
type Memory struct {
	values      map[string]string
	unavailable bool
}

// NewMemory returns an empty in-memory store.
func NewMemory() *Memory { return &Memory{values: map[string]string{}} }

// SetUnavailable makes every operation fail, to test the failure path.
func (m *Memory) SetUnavailable(v bool) { m.unavailable = v }

func (m *Memory) Get(key string) (string, error) {
	if m.unavailable {
		return "", ErrUnavailable
	}
	v, ok := m.values[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (m *Memory) Set(key, value string) error {
	if m.unavailable {
		return ErrUnavailable
	}
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("refusing to store an empty credential")
	}
	m.values[key] = value
	return nil
}

func (m *Memory) Delete(key string) error {
	if m.unavailable {
		return ErrUnavailable
	}
	delete(m.values, key)
	return nil
}

func (m *Memory) Available() bool { return !m.unavailable }
