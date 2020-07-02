package webhooksecret

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1alpha1 "github.com/bigkevmcd/webhook-secret-operator/pkg/apis/apps/v1alpha1"
	"github.com/bigkevmcd/webhook-secret-operator/pkg/git"
)

var log = logf.Log.WithName("controller_webhooksecret")

// Add creates a new WebhookSecret Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	cf := git.NewClientFactory(git.NewDriverIdentifier())
	return &ReconcileWebhookSecret{
		kubeClient:       mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		secretFactory:    &secretFactory{stringGenerator: generateSecureString},
		gitClientFactory: cf,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("webhooksecret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1alpha1.WebhookSecret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.WebhookSecret{},
	})
	if err != nil {
		return err
	}
	return nil
}

// ReconcileWebhookSecret reconciles a WebhookSecret object
type ReconcileWebhookSecret struct {
	kubeClient       client.Client
	scheme           *runtime.Scheme
	secretFactory    *secretFactory
	gitClientFactory git.ClientFactory

	routeGetter RouteGetter
}

// Reconcile reads that state of the cluster for a WebhookSecret object and makes changes based on the state read
// and what is in the WebhookSecret.Spec
func (r *ReconcileWebhookSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling WebhookSecret")

	// Fetch the WebhookSecret instance
	instance := &v1alpha1.WebhookSecret{}
	err := r.kubeClient.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.reconcileDelete(reqLogger, request)
		}
		return reconcile.Result{}, err
	}
	secret, err := r.secretFactory.CreateSecret(instance)
	if err := controllerutil.SetControllerReference(instance, secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &corev1.Secret{}
	err = r.kubeClient.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// TODO: this also needs to check for the webhook state.
		return r.reconcileNewSecret(reqLogger, instance, secret)
	} else if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Skip reconcile: Secret already exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{}, nil
}

func (r *ReconcileWebhookSecret) reconcileDelete(log logr.Logger, request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// TODO: improve the error messages.
func (r *ReconcileWebhookSecret) reconcileNewSecret(log logr.Logger, w *v1alpha1.WebhookSecret, s *corev1.Secret) (reconcile.Result, error) {
	// TODO: split this up into creating the secret, and updating the git host.
	log.Info("Creating a new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
	err := r.kubeClient.Create(context.TODO(), s)
	if err != nil {
		return reconcile.Result{}, err
	}
	w.Status.SecretRef = v1alpha1.WebhookSecretRef{
		Name: s.Name,
	}
	err = r.kubeClient.Status().Update(context.TODO(), w)
	if err != nil {
		log.Error(err, "Failed to update WebhookSecret status")
		return reconcile.Result{}, err
	}
	client, err := r.gitClientFactory.Create(w.Spec.RepoURL, os.Getenv("GITHUB_AUTH_TOKEN"))
	if err != nil {
		return reconcile.Result{}, err
	}
	// TODO: work out how to get the secret string without having to grab it from the
	// Data.
	repo, err := repoFromURL(w.Spec.RepoURL)
	if err != nil {
		return reconcile.Result{}, err
	}
	hookID, err := client.Create(context.Background(), repo, "https://example.com", string(s.Data["token"]))
	if err != nil {
		return reconcile.Result{}, err
	}
	w.Status.WebhookID = hookID
	err = r.kubeClient.Status().Update(context.TODO(), w)
	if err != nil {
		log.Error(err, "Failed to update WebhookSecret status")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// TODO: move this to the git package.
func repoFromURL(s string) (string, error) {
	parsed, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSuffix(parsed.Path, ".git"), "/"), nil
}
