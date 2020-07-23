# webhook-secret-operator

This is an operator that creates and manages secrets between GitHub/GitLab and your local cluster.

```yaml
apiVersion: apps.bigkevmcd.com/v1alpha1
kind: WebhookSecret
metadata:
  name: example-webhooksecret
spec:
  repoURL: https://github.com/my-org/gitops.git
  secretRef:
    name: "test-secret"
  authSecretRef:
    name: "test-hooks-secret"
  webhookURL:
    hookURL: "https://example.com/"
```

This Kubernetes object creates a secret called `test-secret`, then creates a webhook in the repo `https://github.com/my-org/gitops.git`, pointing at `https://example.com`.

To authenticate the request, the secret in `authSecretRef` is used.

## Installation

You'll need to install the operator first.

```shell
$ kubectl apply -f https://github.com/bigkevmcd/tekton-polling-operator/releases/download/v0.2.0/release-v0.2.0.yaml
```

## Creating a Secret to authenticate with your Git provider



