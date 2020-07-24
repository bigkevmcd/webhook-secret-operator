package git

import (
	"fmt"
	"net/url"
	"strings"

	v1alpha1 "github.com/bigkevmcd/webhook-secret-operator/pkg/apis/apps/v1alpha1"
	scmfactory "github.com/jenkins-x/go-scm/scm/factory"
)

// SCMHooksClientFactory is an implementation of the GitClientFactory interface that can
// create clients based on go-scm.
type SCMHooksClientFactory struct {
	drivers DriverIdentifier
}

// NewClientFactory creates and returns an SCMHookClientFactory.
func NewClientFactory(d DriverIdentifier) *SCMHooksClientFactory {
	return &SCMHooksClientFactory{drivers: d}
}

func (s *SCMHooksClientFactory) ClientForRepo(repo v1alpha1.Repo, token string) (HooksClient, error) {
	// TODO: this should DEBUG log out the identification for URLs.
	driver, err := s.drivers.Identify(repo.URL)

	if err != nil && !IsUnknownDriver(err) {
		return nil, err
	}

	if err != nil && repo.Driver == "" {
		return nil, err
	}

	if repo.Driver != "" {
		driver = repo.Driver
	}

	endpoint := ""
	if repo.Endpoint != "" {
		endpoint = repo.Endpoint
	}

	scmClient, err := scmfactory.NewClient(driver, endpoint, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create a git driver: %s", err)
	}
	r, err := repoFromURL(repo.URL)
	if err != nil {
		return nil, err
	}
	return New(scmClient, r), nil
}

func repoFromURL(s string) (string, error) {
	parsed, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("failed to parse repo from URL %#v: %s", s, err)
	}
	return strings.TrimPrefix(strings.TrimSuffix(parsed.Path, ".git"), "/"), nil
}
