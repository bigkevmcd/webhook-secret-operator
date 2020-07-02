package webhooksecret

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
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
)

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
		gitClientFactory: &stubClientFactory{client: newStubHookClient(t, testWebhookID)},
	}
}

func TestWebhookSecretController(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	ws := makeWebhookSecret()
	cl, r := makeReconciler(t, ws, ws)

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
	r.gitClientFactory.(*stubClientFactory).client.assertHookCreated(testRepo, testHookEndpoint, stubSecret)
}

func makeWebhookSecret() *v1alpha1.WebhookSecret {
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
			WebhookURLRef: v1alpha1.HookRouteRef{
				Route: &v1alpha1.Reference{
					Name:      "my-test-route",
					Namespace: "route-test",
				},
			},
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

type stubClientFactory struct {
	client *stubHookClient
}

func (s stubClientFactory) Create(url, token string) (git.Hooks, error) {
	return s.client, nil
}

func newStubHookClient(t *testing.T, s string) *stubHookClient {
	return &stubHookClient{
		hookID:  s,
		created: make(map[string]string),
		t:       t,
	}
}

var _ HookClient = (*stubHookClient)(nil)

type stubHookClient struct {
	t       *testing.T
	hookID  string
	created map[string]string
}

// TODO: This should do some "authentication" on the creation.
func (s *stubHookClient) Create(ctx context.Context, repo, repoURL, secret string) (string, error) {
	s.created[key(repo, repoURL)] = secret
	return s.hookID, nil
}

func (s *stubHookClient) assertHookCreated(repo, hookURL, wantSecret string) {
	secret := s.created[key(repo, hookURL)]
	if secret != wantSecret {
		s.t.Fatalf("hook creation failed: got %#v, want %#v", secret, wantSecret)
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
