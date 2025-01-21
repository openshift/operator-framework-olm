package olm

import (
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/olm/plugins"
)

func init() {
	operatorPlugInFactoryFuncs = plugins.OperatorPlugInFactoryMap{
		// labels unlabeled non-payload openshift-* csv namespaces with
		// security.openshift.io/scc.podSecurityLabelSync: true
		CsvLabelerPluginID: plugins.NewCsvNamespaceLabelerPluginFunc,
	}
}

func IsPluginEnabled(pluginID plugins.PluginID) bool {
	_, ok := operatorPlugInFactoryFuncs[pluginID]
	return ok
}
