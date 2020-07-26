package test

import (
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// MakeRoute is a test helper that creates routes.
func MakeRoute(id types.NamespacedName) *routev1.Route {
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
