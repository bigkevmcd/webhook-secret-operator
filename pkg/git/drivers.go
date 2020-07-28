package git

import (
	"fmt"
	"net/url"
)

// URLDriverIdentifier is an implementation of the DriverIdentifier interface
// that looks up hosts in a map.
type URLDriverIdentifier struct {
	hosts map[string]string
}

func (u *URLDriverIdentifier) Identify(repoURL string) (string, error) {
	parsed, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse the repoURL %q: %w", repoURL, err)
	}
	d, ok := u.hosts[parsed.Host]
	if ok {
		return d, nil
	}
	return "", unknownDriverError{url: repoURL}
}

// NewDriverIdentifier creates and returns a new URLDriverIdentifier.
func NewDriverIdentifier() *URLDriverIdentifier {
	return &URLDriverIdentifier{
		hosts: map[string]string{
			"github.com": "github",
			"gitlab.com": "gitlab",
		},
	}
}

type unknownDriverError struct {
	url string
}

func (e unknownDriverError) Error() string {
	return fmt.Sprintf("unable to identify driver from URL: %s", e.url)
}

// IsUnknownDriver returns true if the provided error means that we couldn't
// identify the driver from a Repo URL.
func IsUnknownDriver(err error) bool {
	_, ok := err.(unknownDriverError)
	return ok
}
