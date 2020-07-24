package git

import (
	"context"

	v1alpha1 "github.com/bigkevmcd/webhook-secret-operator/pkg/apis/apps/v1alpha1"
)

// ClientFactory is an interface for creating SCM clients based on the URL
// to be fetched.
type ClientFactory interface {
	// ClientForRepo creates a new client, using the provided token for authentication.
	ClientForRepo(repo v1alpha1.Repo, token string) (HooksClient, error)
}

// DriverIdentifer parses a URL and attempts to determine which go-scm driver to
// use to talk to the server.
type DriverIdentifier interface {
	Identify(url string) (string, error)
}

// HooksClient is the API for managing hooks.
type HooksClient interface {
	// Create creates a new repository webhook.
	Create(ctx context.Context, hookURL, secret string) (string, error)

	// Delete deletes a repository webhook.
	Delete(ctx context.Context, hookID string) error
}
