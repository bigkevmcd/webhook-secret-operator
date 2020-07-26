package routes

import (
	"context"
	"fmt"
	"net/url"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KubeRouteGetter is an implementation of RouteGetter.
type KubeRouteGetter struct {
	kubeClient client.Client
}

// New creates and returns a KubeRouteGetter that looks up Routes in k8s.
func New(c client.Client) *KubeRouteGetter {
	return &KubeRouteGetter{
		kubeClient: c,
	}
}

// RouteURL looks for a namespaced Route, and returns the URL from
// it, or an error if not found.
func (k KubeRouteGetter) RouteURL(ctx context.Context, id types.NamespacedName, path string) (string, error) {
	loaded := &routev1.Route{}
	err := k.kubeClient.Get(context.TODO(), id, loaded)
	if err != nil {
		return "", fmt.Errorf("error getting route %s/%s: %w", id.Namespace, id.Name, err)
	}

	scheme := "http"
	if loaded.Spec.TLS != nil {
		scheme = "https"
	}
	if path == "" {
		path = "/"
	}
	routeURL := url.URL{
		Scheme: scheme,
		Host:   loaded.Spec.Host,
		Path:   path,
	}
	return routeURL.String(), err
}
