package git

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
)

// New creates and returns a new SCMHookClient.
func New(c *scm.Client) *SCMHookClient {
	return &SCMHookClient{Client: c}
}

// SCMHookClient is a wrapper for the go-scm scm.Client with a simplified API.
type SCMHookClient struct {
	Client *scm.Client
}

// TODO: this shoudl accept a logr and log out creations.

// Create creates a new repository webhook.
//
// If an HTTP error is returned by the upstream service, an error with the
// response status code is returned.
func (c *SCMHookClient) Create(ctx context.Context, repo, hookURL, secret string) (string, error) {
	hook, r, err := c.Client.Repositories.CreateHook(ctx, repo,
		&scm.HookInput{Target: hookURL, Secret: secret, Events: scm.HookEvents{Push: true}})
	if r != nil && isErrorStatus(r.Status) {
		return "", SCMError{msg: fmt.Sprintf("failed to create hook in repo %s", repo), Status: r.Status}
	}
	if err != nil {
		return "", err
	}
	return hook.ID, nil
}

func isErrorStatus(i int) bool {
	return i >= 400
}
