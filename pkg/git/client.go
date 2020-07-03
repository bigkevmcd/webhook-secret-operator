package git

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
)

// New creates and returns a new SCMHooksClient.
func New(c *scm.Client, r string) *SCMHooksClient {
	return &SCMHooksClient{Client: c, Repo: r}
}

// SCMHooksClient is a wrapper for the go-scm scm.Client with a simplified API.
type SCMHooksClient struct {
	Client *scm.Client
	Repo   string
}

// TODO: this should accept a logr and log out creations.

// Create creates a new repository webhook.
//
// If an HTTP error is returned by the upstream service, an error with the
// response status code is returned.
func (c *SCMHooksClient) Create(ctx context.Context, hookURL, secret string) (string, error) {
	hook, r, err := c.Client.Repositories.CreateHook(ctx, c.Repo,
		&scm.HookInput{Target: hookURL, Secret: secret, Events: scm.HookEvents{Push: true}})
	if r != nil && isErrorStatus(r.Status) {
		return "", SCMError{msg: fmt.Sprintf("failed to create hook in repo %s", c.Repo), Status: r.Status}
	}
	if err != nil {
		return "", err
	}
	return hook.ID, nil
}

// Delete removes a repository webhook.
//
// If an HTTP error is returned by the upstream service, an error with the
// response status code is returned.
func (c *SCMHooksClient) Delete(ctx context.Context, hookID string) error {
	r, err := c.Client.Repositories.DeleteHook(ctx, c.Repo, hookID)
	if r != nil && isErrorStatus(r.Status) {
		return SCMError{msg: fmt.Sprintf("failed to create hook in repo %s", c.Repo), Status: r.Status}
	}
	if err != nil {
		return err
	}
	return nil
}

func isErrorStatus(i int) bool {
	return i >= 400
}
