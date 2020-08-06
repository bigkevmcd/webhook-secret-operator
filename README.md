# webhook-secret-operator ![Go](https://github.com/bigkevmcd/webhook-secret-operator/workflows/Go/badge.svg)

This is an operator that creates and manages secrets between GitHub/GitLab and your local cluster.

**NOTE**: This is a very early release of this code.

```yaml
apiVersion: apps.bigkevmcd.com/v1alpha1
kind: WebhookSecret
metadata:
  name: example-webhooksecret
spec:
  repo: 
    url: https://github.com/my-org/gitops.git
  authSecretRef:
    name: demo-hooks-secret
  webhookURL:
    hookURL: https://example.com/
```

This Kubernetes object creates a Secret called `example-webhooksecret`, then creates a webhook in the repo `https://github.com/my-org/gitops.git`, pointing at `https://example.com`, and sharing the token from the `example-webhooksecret` Secret.

To authenticate the request, the secret in `authSecretRef`, `demo-hooks-secret` is used.

## Installation

You'll need to install the operator first.

```shell
$ kubectl apply -f https://github.com/bigkevmcd/webhook-secret-operator/releases/download/v0.4.1/release-v0.4.1.yaml
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
  repo: 
    url: https://github.com/my-org/gitops.git
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
  repo:
    url: https://github.com/my-org/gitops.git
  authSecretRef:
    name: demo-hooks-secret
  webhookURL:
    routeRef:
      name: name-of-route
      namespace: route-ns
```

This will calculate the URL for the route and populate the hook URL with it.

### Configuring the key name within the secret

By default, the secret will be generated and placed into the `token` key within
the generated secret.

If you want to override this, add a `key` to the spec:

```yaml
apiVersion: apps.bigkevmcd.com/v1alpha1
kind: WebhookSecret
metadata:
  name: example-webhooksecret
spec:
  repo:
    url: https://github.com/my-org/gitops.git
  key: not-token
  authSecretRef:
    name: demo-hooks-secret
  webhookURL:
    routeRef:
      name: name-of-route
      namespace: route-ns
```
