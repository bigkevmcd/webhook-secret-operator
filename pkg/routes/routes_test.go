package routes

import (
	"context"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ RouteGetter = (*KubeRouteGetter)(nil)

var testID = types.NamespacedName{Name: "test-route", Namespace: "test-ns"}

func TestRouteURL(t *testing.T) {
	g := makeGetter(createRoute(testID))

	hookURL, err := g.RouteURL(context.TODO(), testID)
	if err != nil {
		t.Fatal(err)
	}

	if hookURL != "https://test.example.com" {
		t.Fatalf("got %s, want 'https://test.example.com/", hookURL)
	}
}

func TestRouteURLWithMissingRoute(t *testing.T) {
	g := makeGetter()

	_, err := g.RouteURL(context.TODO(), testID)

	if err.Error() != `error getting route test-ns/test-route: routes.route.openshift.io "test-route" not found` {
		t.Fatal(err)
	}
}

func createRoute(id types.NamespacedName) *routev1.Route {
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      id.Name,
			Namespace: id.Namespace,
		},
		Spec: routev1.RouteSpec{
			Host: "test.example.com",
			TLS:  &routev1.TLSConfig{},
		},
	}
}

func makeGetter(o ...runtime.Object) *KubeRouteGetter {
	s := scheme.Scheme
	routev1.AddToScheme(s)
	return New(fake.NewFakeClient(o...))
}
