#! /bin/bash

set -o pipefail
set -o nounset
set -o errexit

KUBEBUILDER_BIN=${KUBEBUILDER_BIN:=/usr/local/kubebuilder}
OS=$(go env GOOS)
ARCH=$(go env GOARCH)
KUBEBUILDER_RELEASE=2.3.1
KUBEBUILDER_DOWNLOAD_URL="https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_RELEASE}/kubebuilder_${KUBEBUILDER_RELEASE}_${OS}_${ARCH}.tar.gz"

# Note: From v3.0.0+ the release assets in the kubebuilder repo change format
# See: https://github.com/kubernetes-sigs/kubebuilder/releases
# KUBEBUILDER_DOWNLOAD_URL="https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_RELEASE}/kubebuilder_${OS}_${ARCH}"

env

if [[ -d ${KUBEBUILDER_BIN} ]]; then
    echo "Not installing kubebuilder as the binary already exists in \$PATH"
    exit 0
fi

curl -L ${KUBEBUILDER_DOWNLOAD_URL} | tar -xz -C /tmp/ && \
    mv /tmp/kubebuilder_${KUBEBUILDER_RELEASE}_${OS}_${ARCH}/ /usr/local/kubebuilder

echo "Kubebuilder installation complete!"
echo "Run the following locally to ensure kubebuilder is added to \$PATH"
echo "export PATH=$PATH:/usr/local/kubebuilder/bin"
