#! /bin/bash

set -o pipefail
set -o nounset
set -o errexit

KUBEBUILDER_BIN=${KUBEBUILDER_BIN:=/usr/local/kubebuilder}
OS=$(go env GOOS)
ARCH=$(go env GOARCH)
KUBEBUILDER_RELEASE=2.3.1

if [[ -d ${KUBEBUILDER_BIN} ]]; then
    echo "Not installing kubebuilder as the binary already exists in \$PATH"
    exit 0
fi

curl -L "https://go.kubebuilder.io/dl/${KUBEBUILDER_RELEASE}/${OS}/${ARCH}" | tar -xz -C /tmp/ && \
    mv /tmp/kubebuilder_${KUBEBUILDER_RELEASE}_${OS}_${ARCH}/ /usr/local/kubebuilder

echo "Kubebuilder installation complete!"
echo "Run the following locally to ensure kubebuilder is added to \$PATH"
echo "export PATH=$PATH:/usr/local/kubebuilder/bin"
