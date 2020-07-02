package git

import (
	"testing"

	"github.com/bigkevmcd/webhook-secret-operator/test"
	"github.com/jenkins-x/go-scm/scm"
)

var _ ClientFactory = (*SCMHooksClientFactory)(nil)

func TestSCMFactory(t *testing.T) {
	// TODO non-standard GitHub and GitLab hosts!
	// Probably need to return the serverURL part for the scm factory too.
	urlTests := []struct {
		gitURL  string
		want    scm.Driver
		wantErr string
	}{
		{"https://github.com/myorg/myrepo.git", scm.DriverGithub, ""},
		{"https://gitlab.com/myorg/myrepo/myother.git", scm.DriverGitlab, ""},
		{"https://scm.example.com/myorg/myother.git", scm.DriverUnknown, "unable to identify driver"},
	}
	factory := NewClientFactory(NewDriverIdentifier())
	for _, tt := range urlTests {
		t.Run(tt.gitURL, func(rt *testing.T) {
			client, err := factory.Create(tt.gitURL, "test-token")
			if !test.MatchError(rt, tt.wantErr, err) {
				rt.Errorf("error failed to match, got %#v, want %s", err, tt.wantErr)
			}
			if tt.want == scm.DriverUnknown {
				return
			}
			gc, ok := client.(*SCMHooksClient)
			if !ok {
				rt.Errorf("returned client is not an SCMHooksClient: %T", gc)
			}
			if gc.Client.Driver != tt.want {
				rt.Errorf("got %s, want %s", gc.Client.Driver, tt.want)
			}
		})
	}

}
