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
)

func makeReconciler(ws *v1alpha1.WebhookSecret, objs ...runtime.Object) (client.Client, *ReconcileWebhookSecret) {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, ws)
	cl := fake.NewFakeClient(objs...)
	return cl, &ReconcileWebhookSecret{
		kubeClient: cl,
		scheme:     s,
		secretFactory: &secretFactory{
			stringGenerator: func() (string, error) {
				return "known-secret", nil
			},
		},
		gitClientFactory: &stubClientFactory{client: newStubHookClient(testWebhookID)},
	}
}

func TestWebhookSecretController(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	ws := makeWebhookSecret()
	cl, r := makeReconciler(ws, ws)

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

}

func makeWebhookSecret() *v1alpha1.WebhookSecret {
	return &v1alpha1.WebhookSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testWebhookSecretName,
			Namespace: testWebhookSecretNamespace,
		},
		Spec: v1alpha1.WebhookSecretSpec{
			SecretRef: v1alpha1.WebhookSecretRef{
				Name: testSecretName,
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

func newStubHookClient(s string) *stubHookClient {
	return &stubHookClient{
		hookID:  s,
		created: make(map[string]string),
	}
}

type stubHookClient struct {
	hookID  string
	created map[string]string
}

func (s *stubHookClient) Create(ctx context.Context, repo, repoURL, secret string) (string, error) {
	s.created[key(repo, repoURL)] = secret
	return s.hookID, nil
}

func key(s ...string) string {
	return strings.Join(s, ":")
}
