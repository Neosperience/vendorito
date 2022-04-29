package vendorito

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/containers/image/v5/types"
)

var (
	ErrRegistryNotFound = errors.New("registry not found")
	ErrUserNotFound     = errors.New("user not found")
)

// ParseCredentials parses a list of domain:user:pass pairs and returns an AuthStore
func ParseCredentials(credentials []string) (*AuthStore, error) {
	store := NewAuthStore()
	for _, cred := range credentials {
		parts := strings.SplitN(cred, ":", 3)
		if len(parts) != 3 {
			return nil, errors.New("invalid credentials format")
		}
		store.Add(parts[0], parts[1], parts[2])
	}
	return store, nil
}

// AddAuthToContext adds the auth credentials to the context either from the URL or from the store.
// If "mustAuth" is true, the function will return an error if no credentials are found.
func AddAuthToContext(context *types.SystemContext, authStore *AuthStore, info *url.URL, mustAuth bool) error {
	if info.User != nil {
		password, ok := info.User.Password()

		// If we only specified the user, get the password from auth store
		if !ok {
			authConfig, err := authStore.Get(info.Host, info.User.Username())
			if err != nil {
				return fmt.Errorf("could not get password for user %s (%s): %w", info.User.Username(), info.Host, err)
			}
			context.DockerAuthConfig = authConfig
		} else {
			context.DockerAuthConfig = &types.DockerAuthConfig{
				Username: info.User.Username(),
				Password: password,
			}
		}
		return nil
	}

	// No user specified, try to get the default user from the auth store, where available
	user, err := authStore.Get(info.Host, "")
	if err != nil {
		// No default user found
		if !mustAuth && (err == ErrUserNotFound || err == ErrRegistryNotFound) {
			// We don't need to authenticate, so we can just exit
			return nil
		}
		return fmt.Errorf("could not get default user for registry %s: %w", info.Host, err)
	}
	context.DockerAuthConfig = user
	return nil
}

// AuthStore is a store for authentication credentials to docker registries.
type AuthStore struct {
	users map[string]map[string]*types.DockerAuthConfig
}

// NewAuthStore creates a new AuthStore.
func NewAuthStore() *AuthStore {
	return &AuthStore{
		users: make(map[string]map[string]*types.DockerAuthConfig),
	}
}

// Add adds a new user:pass pair to the store for a registry.
func (s *AuthStore) Add(registry string, user string, password string) {
	if _, ok := s.users[registry]; !ok {
		s.users[registry] = make(map[string]*types.DockerAuthConfig)
	}
	s.users[registry][user] = &types.DockerAuthConfig{
		Username: user,
		Password: password,
	}
}

// Get returns the user:pass pair for a registry for either a given user or for the first user available (if user is empty).
func (s *AuthStore) Get(registry string, user string) (*types.DockerAuthConfig, error) {
	users, ok := s.users[registry]
	if !ok {
		return nil, ErrRegistryNotFound
	}

	// If user is empty, return the first user available.
	if user == "" {
		for _, u := range users {
			return u, nil
		}
	}

	// Return the user if it exists.
	userData, ok := users[user]
	if !ok {
		return nil, ErrUserNotFound
	}
	return userData, nil
}
