package webhooksecret

import (
	"context"
	"fmt"
	syslog "log"

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
	"github.com/bigkevmcd/webhook-secret-operator/pkg/secrets"
)

var log = logf.Log.WithName("controller_webhooksecret")

const webhookFinalizer = "webhooksecrets.finalizer"

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
		authSecretGetter: secrets.New(mgr.GetClient()),
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

	authSecretGetter secrets.SecretGetter
	routeGetter      RouteGetter
}

// Reconcile reads that state of the cluster for a WebhookSecret object and makes changes based on the state read
// and what is in the WebhookSecret.Spec
func (r *ReconcileWebhookSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling WebhookSecret")

	ctx := context.Background()

	// Fetch the WebhookSecret instance
	instance := &v1alpha1.WebhookSecret{}
	err := r.kubeClient.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(instance.ObjectMeta.Finalizers, webhookFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, webhookFinalizer)
			if err := r.kubeClient.Update(ctx, instance); err != nil {
				return reconcile.Result{}, fmt.Errorf("failed to add the finalizer: %s", err)
			}
		}
	} else {
		if containsString(instance.ObjectMeta.Finalizers, webhookFinalizer) {
			if err := r.deleteWebhook(ctx, instance); err != nil {
				return reconcile.Result{}, fmt.Errorf("failed to delete the webhook: %s", err)
			}
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, webhookFinalizer)
			if err := r.kubeClient.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, fmt.Errorf("failed to update the finalizers after deleting the webhook: %s", err)
			}
		}
		return reconcile.Result{}, nil
	}

	secret, err := r.secretFactory.CreateSecret(instance)
	if err := controllerutil.SetControllerReference(instance, secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &corev1.Secret{}
	err = r.kubeClient.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// TODO: this also needs to check for the webhook state.
		return r.reconcileNewSecret(ctx, reqLogger, instance, secret)
	} else if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Skip reconcile: Secret already exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{}, nil
}

func (r *ReconcileWebhookSecret) authenticatedClient(ctx context.Context, ws *v1alpha1.WebhookSecret) (git.HooksClient, error) {
	authToken, err := r.authSecretGetter.SecretToken(ctx, types.NamespacedName{Name: ws.Spec.AuthSecretRef.Name, Namespace: ws.ObjectMeta.Namespace})
	if err != nil {
		log.Error(err, "failed to get the authentication token")
		return nil, fmt.Errorf("could not get authentication token from %s/%s: %s", ws.Spec.AuthSecretRef.Name, ws.ObjectMeta.Namespace, err)
	}
	client, err := r.gitClientFactory.ClientForRepo(ws.Spec.RepoURL, authToken)
	if err != nil {
		return nil, fmt.Errorf("could not get client from %s: %s", ws.Spec.RepoURL, err)
	}
	return client, nil
}

func (r *ReconcileWebhookSecret) deleteWebhook(ctx context.Context, ws *v1alpha1.WebhookSecret) error {
	client, err := r.authenticatedClient(ctx, ws)
	if err != nil {
		return err
	}
	err = client.Delete(ctx, ws.Status.WebhookID)
	if err != nil {
		return err
	}
	return nil
}

// TODO: improve the error messages.
func (r *ReconcileWebhookSecret) reconcileNewSecret(ctx context.Context, log logr.Logger, ws *v1alpha1.WebhookSecret, s *corev1.Secret) (reconcile.Result, error) {
	// TODO: split this up into creating the secret, and updating the git host.
	log.Info("Creating a new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
	err := r.kubeClient.Create(ctx, s)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to create a secret: %s", err)
	}
	ws.Status.SecretRef = v1alpha1.WebhookSecretRef{
		Name: s.Name,
	}
	err = r.kubeClient.Status().Update(ctx, ws)
	if err != nil {
		log.Error(err, "failed to update WebhookSecret status")
		return reconcile.Result{}, fmt.Errorf("failed to update status after creating a secret: %s", err)
	}

	// TODO: work out how to get the secret string without having to grab it from the
	// Data.
	hookID, err := r.createWebhook(ctx, ws, string(s.Data["token"]))
	if err != nil {
		return reconcile.Result{}, err
	}
	ws.Status.WebhookID = hookID
	err = r.kubeClient.Status().Update(ctx, ws)
	if err != nil {
		log.Error(err, "Failed to update WebhookSecret status")
		return reconcile.Result{}, fmt.Errorf("failed to update status after creating a Webhook: %s", err)
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileWebhookSecret) createWebhook(ctx context.Context, ws *v1alpha1.WebhookSecret, secret string) (string, error) {
	hookURL, err := r.hookURL(ws.Spec.WebhookURL)
	if err != nil {
		log.Error(err, "Failed to get the URL for route")
		return "", err
	}

	client, err := r.authenticatedClient(ctx, ws)
	hookID, err := client.Create(ctx, hookURL, secret)
	if err != nil {
		return "", err
	}
	return hookID, nil
}

func (r *ReconcileWebhookSecret) hookURL(u v1alpha1.HookRoute) (string, error) {
	if u.HookURL != "" {
		return u.HookURL, nil
	}
	hookURL, err := r.routeGetter.RouteURL(u.RouteRef.NamespacedName())
	if err != nil {
		log.Error(err, "Failed to get the URL for route")
		return "", err
	}
	if u.RouteRef.Path != "" {
		hookURL = hookURL + u.RouteRef.Path
	}
	syslog.Printf("KEVIN!!!! returning %#v\n", hookURL)
	return hookURL, nil
}
