package webhooksecret

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	v1alpha1 "github.com/bigkevmcd/webhook-secret-operator/pkg/apis/apps/v1alpha1"
	"github.com/bigkevmcd/webhook-secret-operator/pkg/git"
	"github.com/bigkevmcd/webhook-secret-operator/pkg/secrets"
)

const (
	testWebhookSecretName      = "test-webhook-secret"
	testWebhookSecretNamespace = "test-webhook-ns"
	testSecretName             = "test-secret"
	testWebhookID              = "1234567"
	testHookEndpoint           = "https://example.com/test"
	testRepoURL                = "https://github.com/example/example.git"
	testRepo                   = "example/example"
	stubSecret                 = "known-secret"
	testAuthSecretName         = "auth-secret"
	testAuthToken              = "test-auth-token"
)

// Reconciling a simple WebhookSecret should create a Secret, and create a
// webhook in the repository pointing to the HookURL in the WebhookSecret.
func TestWebhookSecretControllerWithAHookURL(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	ws := makeWebhookSecret(v1alpha1.HookRoute{
		HookURL: testHookEndpoint,
	})
	cl, r := makeReconciler(t, ws, ws, makeTestSecret(testAuthSecretName))
	req := makeReconcileRequest()

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Requeue {
		t.Fatal("request was requeued")
	}
	s := &corev1.Secret{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: ws.Spec.SecretRef.Name, Namespace: req.Namespace}, s)
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	ws = &v1alpha1.WebhookSecret{}
	err = r.kubeClient.Get(context.TODO(), req.NamespacedName, ws)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Status.SecretRef.Name != s.Name {
		t.Fatalf("got incorrect secret in status, got %#v, want %#v", ws.Status.SecretRef.Name, s.Name)
	}
	if ws.Status.WebhookID != testWebhookID {
		t.Fatalf("status does not have the correct WebhookID, got %#v, want %#v", ws.Status.WebhookID, testWebhookID)
	}
	r.gitClientFactory.(*stubClientFactory).client.assertHookCreated(testHookEndpoint, stubSecret)
}

// Reconciling a simple WebhookSecret should create a Secret, and create a
// webhook in the repository pointing to the URL of the Route referenced in the
// WebhookSecret.
func TestWebhookSecretControllerWithARouteRef(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	ws := makeWebhookSecret(v1alpha1.HookRoute{
		RouteRef: &v1alpha1.Reference{
			Name:      "my-test-route",
			Namespace: "route-test",
		},
	})
	cl, r := makeReconciler(t, ws, ws, makeTestSecret(testAuthSecretName))
	req := makeReconcileRequest()

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Requeue {
		t.Fatal("request was requeued")
	}
	s := &corev1.Secret{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: ws.Spec.SecretRef.Name, Namespace: req.Namespace}, s)
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	ws = &v1alpha1.WebhookSecret{}
	err = r.kubeClient.Get(context.TODO(), req.NamespacedName, ws)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Status.SecretRef.Name != s.Name {
		t.Fatalf("got incorrect secret in status, got %#v, want %#v", ws.Status.SecretRef.Name, s.Name)
	}
	if ws.Status.WebhookID != testWebhookID {
		t.Fatalf("status does not have the correct WebhookID, got %#v, want %#v", ws.Status.WebhookID, testWebhookID)
	}
	r.gitClientFactory.(*stubClientFactory).client.assertHookCreated(testHookEndpoint, stubSecret)
}

// If the Route referenced does not exist, no hook should be created.
// The WebhookSecret should reflect the error.
func TestWebhookSecretControllerWithARouteRefAndRouteMissing(t *testing.T) {
}

// When reconciling a WebhookSecret, if we have the secret already, but no
// Status.WebhookID, then we should create the webhook.
func TestWebhookSecretControllerSecretButNoHook(t *testing.T) {
	t.Skip()
}

// We're watching the Secret, when it's deleted, we should recreate the Secret,
// and update the Hook's secret (either remove and create, or update).
func TestWebhookSecretDeletedWebhookSecretDeletedSecret(t *testing.T) {
	t.Skip()
}

// When a WebhookSecret is deleted, it should cleanup the webhook in the
// git host.
func TestWebhookSecretControllerDeletedWebhookSecret(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	ws := makeWebhookSecret(v1alpha1.HookRoute{
		HookURL: testHookEndpoint,
	})
	ws.Status.WebhookID = testWebhookID
	ws.ObjectMeta.Finalizers = []string{webhookFinalizer}
	now := metav1.NewTime(time.Now())
	ws.ObjectMeta.DeletionTimestamp = &now

	cl, r := makeReconciler(t, ws, ws, makeTestSecret(testAuthSecretName))
	req := makeReconcileRequest()

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Requeue {
		t.Fatal("request was requeued")
	}
	s := &corev1.Secret{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: ws.Spec.SecretRef.Name, Namespace: req.Namespace}, s)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("secret still exists %v", err)
	}
	r.gitClientFactory.(*stubClientFactory).client.assertHookDeleted(testWebhookID)
}

func makeWebhookSecret(r v1alpha1.HookRoute) *v1alpha1.WebhookSecret {
	return &v1alpha1.WebhookSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testWebhookSecretName,
			Namespace: testWebhookSecretNamespace,
		},
		Spec: v1alpha1.WebhookSecretSpec{
			RepoURL: testRepoURL,
			SecretRef: v1alpha1.WebhookSecretRef{
				Name: testSecretName,
			},
			AuthSecretRef: v1alpha1.WebhookSecretRef{
				Name: testAuthSecretName,
			},
			WebhookURL: r,
		},
	}
}

func makeReconcileRequest() reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      testWebhookSecretName,
			Namespace: testWebhookSecretNamespace,
		},
	}
}

// TODO: move these to the git package as mocks.
type stubClientFactory struct {
	client    *stubHookClient
	authToken string
}

// TODO: ensure that this can fail to find a client.
func (s stubClientFactory) ClientForRepo(url, token string) (git.HooksClient, error) {
	if token != s.authToken {
		return nil, errors.New("failed to authenticate")
	}
	return s.client, nil
}

func newStubHookClient(t *testing.T, repo, hookID string) *stubHookClient {
	return &stubHookClient{
		hookID:  hookID,
		repo:    repo,
		created: make(map[string]string),
		deleted: make(map[string]string),
		t:       t,
	}
}

var _ git.HooksClient = (*stubHookClient)(nil)

type stubHookClient struct {
	t       *testing.T
	repo    string
	hookID  string
	created map[string]string
	deleted map[string]string
}

// TODO: revisit use of s.repo
func (s *stubHookClient) Create(ctx context.Context, hookURL, secret string) (string, error) {
	s.created[key(s.repo, hookURL)] = secret
	return s.hookID, nil
}

func (s *stubHookClient) Delete(ctx context.Context, hookID string) error {
	s.deleted[s.repo] = hookID
	return nil
}

func (s *stubHookClient) assertHookCreated(hookURL, wantSecret string) {
	secret := s.created[key(s.repo, hookURL)]
	if secret != wantSecret {
		s.t.Fatalf("hook creation failed: got %#v, want %#v", secret, wantSecret)
	}
}

func (s *stubHookClient) assertHookDeleted(wantHookID string) {
	deletedID := s.deleted[s.repo]
	if deletedID != wantHookID {
		s.t.Fatalf("hook deletion failed: got %#v, want %#v", deletedID, wantHookID)
	}
}

var _ RouteGetter = (*stubRouteGetter)(nil)

func newStubRouteGetter(s string) *stubRouteGetter {
	return &stubRouteGetter{
		routeURL: s,
	}
}

type stubRouteGetter struct {
	routeURL string
}

// TODO: this should check that the namespaced name matches something.
func (s *stubRouteGetter) RouteURL(types.NamespacedName) (string, error) {
	return s.routeURL, nil
}

func key(s ...string) string {
	return strings.Join(s, ":")
}

func makeTestSecret(n string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: testWebhookSecretNamespace,
		},
		Data: map[string][]byte{
			"token": []byte(testAuthToken),
		},
	}
}

func makeReconciler(t *testing.T, ws *v1alpha1.WebhookSecret, objs ...runtime.Object) (client.Client, *ReconcileWebhookSecret) {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, ws)
	cl := fake.NewFakeClient(objs...)
	return cl, &ReconcileWebhookSecret{
		kubeClient: cl,
		scheme:     s,
		secretFactory: &secretFactory{
			stringGenerator: func() (string, error) {
				return stubSecret, nil
			},
		},
		routeGetter:      newStubRouteGetter(testHookEndpoint),
		gitClientFactory: &stubClientFactory{client: newStubHookClient(t, testRepo, testWebhookID), authToken: testAuthToken},
		authSecretGetter: secrets.New(cl),
	}
}
