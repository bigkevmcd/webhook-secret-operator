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
	g := makeGetter(test.MakeRoute(testID))

	hookURL, err := g.RouteURL(context.TODO(), testID, "")
	if err != nil {
		t.Fatal(err)
	}

	if hookURL != "https://test.example.com/" {
		t.Fatalf("got %s, want 'https://test.example.com/", hookURL)
	}
}

func TestRouteURLWithPath(t *testing.T) {
	g := makeGetter(test.MakeRoute(testID))

	hookURL, err := g.RouteURL(context.TODO(), testID, "/test/api")
	if err != nil {
		t.Fatal(err)
	}

	if hookURL != "https://test.example.com/test/api" {
		t.Fatalf("got %s, want 'https://test.example.com/test/api", hookURL)
	}
}

func TestRouteURLWithMissingRoute(t *testing.T) {
	g := makeGetter()

	_, err := g.RouteURL(context.TODO(), testID, "")

	if err.Error() != `error getting route test-ns/test-route: routes.route.openshift.io "test-route" not found` {
		t.Fatal(err)
	}
}

func makeGetter(o ...runtime.Object) *KubeRouteGetter {
	s := scheme.Scheme
	routev1.AddToScheme(s)
	return New(fake.NewFakeClient(o...))
}
