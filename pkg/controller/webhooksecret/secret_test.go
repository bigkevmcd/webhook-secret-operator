package webhooksecret

import (
	"testing"

	v1alpha1 "github.com/bigkevmcd/webhook-secret-operator/pkg/apis/apps/v1alpha1"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// blank assignment to verify that ReconcileWebhookSecret implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWebhookSecret{}

func TestCreateSecret(t *testing.T) {
	ws := &v1alpha1.WebhookSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test-webhook-secret",
			Namespace: "test-ns",
		},
	}

	want := &corev1.Secret{
		TypeMeta: secretTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test-webhook-secret",
			Namespace: "test-ns",
		},
		Data: map[string][]byte{
			"token": []byte("secret"),
		},
	}
	sf := secretFactory{
		stringGenerator: func() (string, error) {
			return "secret", nil
		},
	}

	secret, err := sf.CreateSecret(ws)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, secret); diff != "" {
		t.Fatalf("incorrect secret generated:\n%s", diff)
	}
}
