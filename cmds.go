package cmds

import (
	_ "github.com/operator-framework/operator-lifecycle-manager/cmd/catalog"
	_ "github.com/operator-framework/operator-lifecycle-manager/cmd/olm"
	_ "github.com/operator-framework/operator-lifecycle-manager/cmd/package-server"
	_ "github.com/operator-framework/operator-lifecycle-manager/util/cpb"

	_ "github.com/grpc-ecosystem/grpc-health-probe"
	_ "github.com/operator-framework/operator-registry/cmd/configmap-server"
	_ "github.com/operator-framework/operator-registry/cmd/initializer"
	_ "github.com/operator-framework/operator-registry/cmd/opm"
	_ "github.com/operator-framework/operator-registry/cmd/registry-server"
)
