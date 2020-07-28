package routes

import (
	"context"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/bigkevmcd/webhook-secret-operator/test"
)

var _ RouteGetter = (*KubeRouteGetter)(nil)

var testID = types.NamespacedName{Name: "test-route", Namespace: "test-ns"}

func TestRouteURL(t *testing.T) {
	g := makeGetter(t, test.MakeRoute(testID))

	hookURL, err := g.RouteURL(context.TODO(), testID, "")
	if err != nil {
		t.Fatal(err)
	}

	if hookURL != "https://example.com/" {
		t.Fatalf("got %s, want 'https://example.com/", hookURL)
	}
}

func TestRouteURLWithPath(t *testing.T) {
	g := makeGetter(t, test.MakeRoute(testID))

	hookURL, err := g.RouteURL(context.TODO(), testID, "/test/api")
	if err != nil {
		t.Fatal(err)
	}

	if hookURL != "https://example.com/test/api" {
		t.Fatalf("got %s, want 'https://example.com/test/api", hookURL)
	}
}

func TestRouteURLWithMissingRoute(t *testing.T) {
	g := makeGetter(t)

	_, err := g.RouteURL(context.TODO(), testID, "")

	if err.Error() != `error getting route test-ns/test-route: routes.route.openshift.io "test-route" not found` {
		t.Fatal(err)
	}
}

func makeGetter(t *testing.T, o ...runtime.Object) *KubeRouteGetter {
	t.Helper()
	s := scheme.Scheme
	if err := routev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return New(fake.NewFakeClient(o...))
}
