package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// kind: WebhookSecret
// apiVersion: webhooks/v1alpha1
// spec:
//   repoURL: "https://github.com/testing/testing.git"
//   authSecretRef:
//     name: "gitops-github-auth-token"
//   secretRef:
//     name: "this-is-the-secret-to-be-created"
//   webhookURLRef:
//     route:
//       name: "el-gitop-eventlistener-route"
//   events:
//     - push
//     - pull_request

// WebhookSecretSpec defines the desired state of WebhookSecret
//
// This is used to authenticate requests to the API for RepoURL.
type WebhookSecretSpec struct {
	RepoURL       string           `json:"repoURL"`
	AuthSecretRef WebhookSecretRef `json:"authSecretRef"`
	SecretRef     WebhookSecretRef `json:"secretRef"`
	WebhookURLRef HookRouteRef     `json:"webhookURLRef"`

	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// WebhookSecretStatus defines the observed state of WebhookSecret
type WebhookSecretStatus struct {
	WebhookID string           `json:"webhookID,omitempty"`
	SecretRef WebhookSecretRef `json:"secretRef,omitempty"`
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// WebhookSecretRef is the secret to be created.
type WebhookSecretRef struct {
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
}

type HookRouteRef struct {
	Route *Reference `json:"routeRef,omitempty"`
}

// Reference is a generic reference with a name/namespace.
type Reference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// NamespacedName returns a NamespacedName for this reference.
func (r Reference) NamespacedName() types.NamespacedName {
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
