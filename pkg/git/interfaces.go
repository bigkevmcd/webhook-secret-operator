package git

import (
	"context"
)

// ClientFactory is an interface for creating SCM clients based on the URL
// to be fetched.
type ClientFactory interface {
	// Create creates a new client, using the provided token for authentication.
	Create(url, token string) (Hooks, error)
}

// DriverIdentifer parses a URL and attempts to determine which go-scm driver to
// use to talk to the server.
type DriverIdentifier interface {
	Identify(url string) (string, error)
}

// Hooks is the API for managing hooks.
type Hooks interface {
	// Create creates a new repository webhook.
	Create(ctx context.Context, repo, hookURL, secret string) (string, error)
}
