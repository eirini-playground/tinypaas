module github.com/jimmykarily/tinypaas/cli

go 1.15

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	k8s.io/client-go => k8s.io/client-go v0.17.6
)

require (
	github.com/pivotal/kpack v0.1.4
	github.com/spf13/cobra v1.1.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)
