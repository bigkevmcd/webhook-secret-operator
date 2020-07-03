package git

import (
	"context"
	"net/http"
	"testing"

	"github.com/bigkevmcd/webhook-secret-operator/test"
	"github.com/jenkins-x/go-scm/scm/factory"
	"gopkg.in/h2non/gock.v1"
)

func TestCreate(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Post("/repos/Codertocat/Hello-World/hooks").
		MatchHeader("Authorization", "Bearer authtoken").
		MatchType("json").
		JSON(map[string]interface{}{
			"name":   "web",
			"active": true,
			"events": []string{"push"},
			"config": map[string]string{
				"url":          "https://example.com/testing",
				"content_type": "json",
				"secret":       "t0ps3cr3t",
			}}).
		Reply(http.StatusCreated).
		Type("application/json").
		File("testdata/hook_created.json")

	scmClient, err := factory.NewClient("github", "", "authtoken")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient, "Codertocat/Hello-World")

	webhookID, err := client.Create(context.TODO(), "https://example.com/testing", "t0ps3cr3t")
	if err != nil {
		t.Fatal(err)
	}
	want := "12345678"
	if webhookID != "12345678" {
		t.Fatalf("got a different WebHookID back: %#v, want %#v", webhookID, want)
	}
}

func TestCreateWithNotFoundResponse(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Post("/repos/Codertocat/Hello-World/hooks").
		MatchHeader("Authorization", "Bearer authtoken").
		MatchType("json").
		JSON(map[string]interface{}{
			"name":   "web",
			"active": true,
			"events": []string{"push"},
			"config": map[string]string{
				"url":          "https://example.com/testing",
				"content_type": "json",
				"secret":       "t0ps3cr3t",
			}}).
		Reply(http.StatusNotFound).
		Type("application/json")

	scmClient, err := factory.NewClient("github", "", "authtoken")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient, "Codertocat/Hello-World")

	_, err = client.Create(context.TODO(), "https://example.com/testing", "t0ps3cr3t")

	if !IsNotFound(err) {
		t.Fatalf("failed with %#v", err)
	}
}

func TestCreateUnableToConnect(t *testing.T) {
	scmClient, err := factory.NewClient("github", "https://localhost:2000", "")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient, "Codertocat/Hello-World")

	_, err = client.Create(context.TODO(), "pipelines.yaml", "master")
	if !test.MatchError(t, "connection refused", err) {
		t.Fatal(err)
	}
}

func TestDelete(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Delete("/repos/Codertocat/Hello-World/hooks/1234567").
		MatchHeader("Authorization", "Bearer authtoken").
		Reply(http.StatusNoContent)

	scmClient, err := factory.NewClient("github", "", "authtoken")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient, "Codertocat/Hello-World")

	err = client.Delete(context.TODO(), "1234567")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteWithNotFoundResponse(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Delete("/repos/Codertocat/Hello-World/hooks/1234567").
		MatchHeader("Authorization", "Bearer authtoken").
		Reply(http.StatusNotFound)

	scmClient, err := factory.NewClient("github", "", "authtoken")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient, "Codertocat/Hello-World")

	err = client.Delete(context.TODO(), "1234567")

	if !IsNotFound(err) {
		t.Fatalf("failed with %#v", err)
	}
}
