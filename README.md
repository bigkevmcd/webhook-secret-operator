# webhook-secret-operator

This is an operator that creates and manages secrets between GitHub/GitLab and your local cluster.

**NOTE**: This is a very early release of this code.

```yaml
apiVersion: apps.bigkevmcd.com/v1alpha1
kind: WebhookSecret
metadata:
  name: example-webhooksecret
spec:
  repoURL: https://github.com/my-org/gitops.git
  secretRef:
    name: test-secret
  authSecretRef:
    name: demo-hooks-secret
  webhookURL:
    hookURL: https://example.com/
```

This Kubernetes object creates a secret called `test-secret`, then creates a webhook in the repo `https://github.com/my-org/gitops.git`, pointing at `https://example.com`.

To authenticate the request, the secret in `authSecretRef`, `demo-hooks-secret`.

## Installation

You'll need to install the operator first.

```shell
$ kubectl apply -f https://github.com/bigkevmcd/webhook-secret-operator/releases/download/v0.2.2/release-v0.2.2.yaml
```

## Creating a Secret to authenticate with your Git provider

You will need to create a secret with an auth token that can create webhooks.

```shell
$ kubectl create secret generic demo-hooks-secret --from-literal=token=<insert a Github Token here>
```

## Automatically creating a webhook secret

### Pointing at a fixed URL

```yaml
apiVersion: apps.bigkevmcd.com/v1alpha1
kind: WebhookSecret
metadata:
  name: example-webhooksecret
spec:
  repoURL: https://github.com/my-org/gitops.git
  secretRef:
    name: test-secret
  authSecretRef:
    name: demo-hooks-secret
  webhookURL:
    hookURL: https://example.com/
```

### Pointing at an OpenShift route

```yaml
apiVersion: apps.bigkevmcd.com/v1alpha1
kind: WebhookSecret
metadata:
  name: example-webhooksecret
spec:
  repoURL: https://github.com/my-org/gitops.git
  secretRef:
    name: test-secret
  authSecretRef:
    name: demo-hooks-secret
  webhookURL:
    routeRef:
      name: name-of-route
      namespace: route-ns
```

This will calculate the URL for the route and populate the hook URL with it.
