package olm

import (
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/olm/plugins"
)

func init() {
	operatorPlugInFactoryFuncs = []plugins.OperatorPlugInFactoryFunc{
		// labels unlabeled non-payload openshift-* csv namespaces with
		// security.openshift.io/scc.podSecurityLabelSync: true
		plugins.NewCsvNamespaceLabelerPluginFunc,
	}
}
