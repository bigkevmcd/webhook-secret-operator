package webhooksecret

import (
	"crypto/rand"

	v1alpha1 "github.com/bigkevmcd/webhook-secret-operator/pkg/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var secretTypeMeta = metav1.TypeMeta{
	Kind:       "Secret",
	APIVersion: "v1",
}

type secretFactory struct {
	stringGenerator func() (string, error)
}

// TODO: this should apply managed-by labels.
func (s *secretFactory) CreateSecret(cr *v1alpha1.WebhookSecret) (*corev1.Secret, error) {
	token, err := s.stringGenerator()
	if err != nil {
		return nil, err
	}
	key := "token"
	if cr.Spec.SecretRef.Key != "" {
		key = cr.Spec.SecretRef.Key
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.SecretRef.Name,
			Namespace: cr.ObjectMeta.Namespace,
		},
		Data: map[string][]byte{
			key: []byte(token),
		},
	}, nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789$:#"

func generateSecureString() (string, error) {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	s := make([]byte, 20)
	for i, v := range b {
		s[i] = charset[int(v)%len(charset)]
	}
	return string(s), nil
}
