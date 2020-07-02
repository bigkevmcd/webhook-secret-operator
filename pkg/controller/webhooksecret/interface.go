package webhooksecret

import (
	"k8s.io/apimachinery/pkg/types"
)

// RouteGetter implementations get the URL for OpenShift Routes.
type RouteGetter interface {
	RouteURL(types.NamespacedName) (string, error)
}
