package git

import (
	"fmt"

	scmfactory "github.com/jenkins-x/go-scm/scm/factory"
)

// SCMHookClientFactory is an implementation of the GitClientFactory interface that can
// create clients based on go-scm.
type SCMHookClientFactory struct {
	drivers DriverIdentifier
}

// NewClientFactory creates and returns an SCMHookClientFactory.
func NewClientFactory(d DriverIdentifier) *SCMHookClientFactory {
	return &SCMHookClientFactory{drivers: d}
}

func (s *SCMHookClientFactory) Create(url, token string) (Hooks, error) {
	// TODO: this should DEBUG log out the identification for URLs.
	driver, err := s.drivers.Identify(url)
	if err != nil {
		return nil, err
	}
	scmClient, err := scmfactory.NewClient(driver, "", token)
	if err != nil {
		return nil, fmt.Errorf("failed to create a git driver: %s", err)
	}
	return New(scmClient), nil
}
