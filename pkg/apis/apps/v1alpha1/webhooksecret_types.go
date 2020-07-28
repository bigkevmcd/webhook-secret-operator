package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// kind: WebhookSecret
// apiVersion: apps.bigkevmcd.com/v1alpha1
// spec:
//   repoURL: "https://github.com/testing/testing.git"
//   authSecretRef:
//     name: "gitops-github-auth-token"
//   webhookURL:
//     route:
//       name: "el-gitop-eventlistener-route"
//   events:
//     - push
//     - pull_request

// WebhookSecretSpec defines the desired state of WebhookSecret
//
// This is used to authenticate requests to the API for Repo.
type WebhookSecretSpec struct {
	Repo          Repo             `json:"repo"`
	AuthSecretRef WebhookSecretRef `json:"authSecretRef"`
	Key           string           `json:"key,omitempty"`
	WebhookURL    HookRoute        `json:"webhookURL"`
}

// WebhookSecretStatus defines the observed state of WebhookSecret
type WebhookSecretStatus struct {
	WebhookID string           `json:"webhookID,omitempty"`
	SecretRef WebhookSecretRef `json:"secretRef,omitempty"`
}

type Repo struct {
	URL      string `json:"url"`
	Driver   string `json:"driver,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

// WebhookSecretRef is the secret to be created.
type WebhookSecretRef struct {
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
}

// HookRoute is the way to get the URL for the Webhook.
//
// HookURL is a static URL.
// RouteRef uses an OpenShift route to calculate the URL.
type HookRoute struct {
	RouteRef *RouteReference `json:"routeRef,omitempty"`
	HookURL  string          `json:"hookURL,omitempty"`
}

// RouteReference is a generic reference with a name/namespace, and the addition
// of a Path to add a custom endpoint.
type RouteReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Path      string `json:"path,omitempty"`
}

// NamespacedName returns a NamespacedName for this reference.
func (r RouteReference) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebhookSecret is the Schema for the webhooksecrets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=webhooksecrets,scope=Namespaced
type WebhookSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebhookSecretSpec   `json:"spec,omitempty"`
	Status WebhookSecretStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebhookSecretList contains a list of WebhookSecret
type WebhookSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebhookSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebhookSecret{}, &WebhookSecretList{})
}
