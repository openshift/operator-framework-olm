module github.com/dweepgogia/new-manifest-verification

go 1.12

require (
	github.com/ghodss/yaml v1.0.0
	github.com/kr/pretty v0.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace github.com/operator-framework/operator-lifecycle-manager => ../operator-lifecycle-manager
