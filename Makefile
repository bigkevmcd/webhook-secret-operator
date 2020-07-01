.PHONY: update-crd
update-crd:
	operator-sdk generate k8s
	operator-sdk generate crds

.PHONY: apply-crd
apply-crd:
	kubectl apply -f deploy/crds/apps.bigkevmcd.com_webhooksecrets_crd.yaml
