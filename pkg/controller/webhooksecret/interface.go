package webhooksecret

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
)

// HookClient implementations provide functionality for creating hooks in a Git
// Hosting Service.
type HookClient interface {
	Create(ctx context.Context, repo, hookURL, secret string) (string, error)
}

// RouteGetter implementations get the URL for OpenShift Routes.
type RouteGetter interface {
	RouteURL(types.NamespacedName) (string, error)
}
