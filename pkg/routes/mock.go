package routes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
)

var _ RouteGetter = (*MockRoute)(nil)

// NewMock returns a simple secret getter.
func NewMock() MockRoute {
	return MockRoute{}
}

// MockRoute implements the RouteGetter interface.
type MockRoute struct {
	secrets map[string]string
}

// RouteURL implements the RouteGetter interface.
func (k MockRoute) RouteURL(ctx context.Context, routeID types.NamespacedName, p string) (string, error) {
	route, ok := k.secrets[key(routeID)]
	if !ok {
		return "", fmt.Errorf("mock not found")
	}
	return route + "/", nil
}

// AddStubResponse is a mock method that sets up a Route to be returned.
func (k MockRoute) AddStubResponse(authToken string, routeID types.NamespacedName, token string) {
	k.secrets[key(routeID)] = token
}

func key(n types.NamespacedName) string {
	return fmt.Sprintf("%s:%s", n.Name, n.Namespace)
}
