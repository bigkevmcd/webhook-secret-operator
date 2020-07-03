package git

import (
	"fmt"
	"net/url"
	"strings"

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

func (s *SCMHooksClientFactory) ClientForRepo(url, token string) (HooksClient, error) {
	// TODO: this should DEBUG log out the identification for URLs.
	driver, err := s.drivers.Identify(url)
	if err != nil {
		return nil, err
	}
	scmClient, err := scmfactory.NewClient(driver, "", token)
	if err != nil {
		return nil, fmt.Errorf("failed to create a git driver: %s", err)
	}
	repo, err := repoFromURL(url)
	if err != nil {
		return nil, err
	}
	return New(scmClient, repo), nil
}

func repoFromURL(s string) (string, error) {
	parsed, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("failed to parse repo from URL %#v: %s", s, err)
	}
	return strings.TrimPrefix(strings.TrimSuffix(parsed.Path, ".git"), "/"), nil
}
