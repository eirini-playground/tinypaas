module github.com/jimmykarily/tinypaas

go 1.15

replace (
	github.com/coreos/bbolt => github.com/coreos/bbolt v1.3.0
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
	k8s.io/client-go => k8s.io/client-go v0.19.2
)

require (
	code.cloudfoundry.org/eirini v0.0.0-20201130204312-0f8caf7ed1aa
	code.cloudfoundry.org/lager v2.0.0+incompatible
	github.com/go-logr/logr v0.3.0
	github.com/pivotal/kpack v0.1.4
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	sigs.k8s.io/controller-runtime v0.6.4
)
