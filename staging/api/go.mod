module github.com/operator-framework/api

go 1.13

require (
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/bshuster-repo/logrus-logstash-hook v1.0.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata/v3 v3.1.3
	github.com/mikefarah/yq/v2 v2.4.1
	github.com/operator-framework/operator-registry v1.5.3
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/controller-tools v0.2.8
)

replace github.com/operator-framework/operator-registry => ../operator-registry
