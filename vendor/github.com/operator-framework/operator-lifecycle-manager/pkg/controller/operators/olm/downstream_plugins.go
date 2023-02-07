package olm

import (
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/olm/plugins"
)

func init() {
	operatorPlugInFactoryFuncs = plugins.OperatorPlugInFactoryMap{
		// labels unlabeled non-payload openshift-* csv namespaces with
		// security.openshift.io/scc.podSecurityLabelSync: true
// TODO: once PSA is enabled downstream, uncomment next line to enable plugin
//		CsvLabelerPluginId: plugins.NewCsvNamespaceLabelerPluginFunc,
	}
}

func IsPluginEnabled(pluginID plugins.PluginID) bool {
	_, ok := operatorPlugInFactoryFuncs[pluginID]
	return ok
}
