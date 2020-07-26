package test

import (
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Host is an option function for MakeRoute to allow configuration fof the
// route.
func Host(h string) routeFunc {
	return func(r *routev1.Route) {
		r.Spec.Host = h
	}
}

type routeFunc func(*routev1.Route)

// MakeRoute is a test helper that creates routes.
func MakeRoute(id types.NamespacedName, opts ...routeFunc) *routev1.Route {
	r := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      id.Name,
			Namespace: id.Namespace,
		},
		Spec: routev1.RouteSpec{
			Host: "example.com",
			TLS:  &routev1.TLSConfig{},
		},
	}
	for _, o := range opts {
		o(r)
	}
	return r
}
