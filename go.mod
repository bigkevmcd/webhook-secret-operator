module github.com/bigkevmcd/webhook-secret-operator

go 1.14

require (
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.4.0
	github.com/jenkins-x/go-scm v1.5.145
	github.com/openshift/api v0.0.0-20200701144905-de5b010b2b38
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/spf13/pflag v1.0.5
	gopkg.in/h2non/gock.v1 v1.0.15
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
