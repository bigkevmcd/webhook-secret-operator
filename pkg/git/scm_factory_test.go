package git

import (
	"testing"

	v1alpha1 "github.com/bigkevmcd/webhook-secret-operator/pkg/apis/apps/v1alpha1"
	"github.com/bigkevmcd/webhook-secret-operator/test"
	"github.com/jenkins-x/go-scm/scm"
)

var _ ClientFactory = (*SCMHooksClientFactory)(nil)

func TestSCMFactory(t *testing.T) {
	// TODO non-standard GitHub and GitLab hosts!
	// Probably need to return the serverURL part for the scm factory too.
	urlTests := []struct {
		repo         v1alpha1.Repo
		wantDriver   scm.Driver
		wantRepo     string
		wantEndpoint string
		wantErr      string
	}{
		{
			repo:       v1alpha1.Repo{URL: "https://github.com/myorg/myrepo.git"},
			wantDriver: scm.DriverGithub,
			wantRepo:   "myorg/myrepo",
			wantErr:    "",
		},
		{
			repo:       v1alpha1.Repo{URL: "https://gitlab.com/myorg/myrepo/myother.git"},
			wantDriver: scm.DriverGitlab,
			wantRepo:   "myorg/myrepo/myother",
			wantErr:    "",
		},
		{
			repo:       v1alpha1.Repo{URL: "https://scm.example.com/myorg/myother.git"},
			wantDriver: scm.DriverUnknown,
			wantRepo:   "",
			wantErr:    "unable to identify driver",
		},
		{
			repo:       v1alpha1.Repo{URL: "https://gitlab.example.com/myorg/myother.git", Driver: "gitlab"},
			wantDriver: scm.DriverGitlab,
			wantRepo:   "myorg/myother",
			wantErr:    "",
		},
		{
			repo:         v1alpha1.Repo{URL: "https://gitlab.example.com/myorg/myother.git", Driver: "gitlab", Endpoint: "https://gitlab.example.com"},
			wantDriver:   scm.DriverGitlab,
			wantRepo:     "myorg/myother",
			wantEndpoint: "https://gitlab.example.com/",
			wantErr:      "",
		},
	}
	factory := NewClientFactory(NewDriverIdentifier())
	for _, tt := range urlTests {
		t.Run(tt.repo.URL, func(rt *testing.T) {
			client, err := factory.ClientForRepo(tt.repo, "test-token")
			if !test.MatchError(rt, tt.wantErr, err) {
				rt.Errorf("error failed to match, got %#v, want %s", err, tt.wantErr)
			}
			if client == nil {
				return
			}
			gc, ok := client.(*SCMHooksClient)
			if !ok {
				rt.Errorf("returned client is not an SCMHooksClient: %T", gc)
			}
			if gc.Client.Driver != tt.wantDriver {
				rt.Errorf("Driver got %s, want %s", gc.Client.Driver, tt.wantDriver)
			}
			if gc.Repo != tt.wantRepo {
				rt.Errorf("Repo got %#v, want %#v", gc.Repo, tt.wantRepo)
			}
			if tt.wantEndpoint != "" && tt.wantEndpoint != gc.Client.BaseURL.String() {
				rt.Errorf("Client BaseURL got %#v, want %#v", gc.Client.BaseURL.String(), tt.wantEndpoint)
			}
		})
	}

}
