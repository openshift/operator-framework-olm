module github.com/dweepgogia/new-manifest-verification

go 1.12

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/spf13/cobra v1.1.1
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.1
)

replace github.com/operator-framework/operator-lifecycle-manager => ../operator-lifecycle-manager
