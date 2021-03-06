# Dockerfile to bootstrap build and test in openshift-ci

FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.16-openshift-4.8

ARG KUBEBUILDER_RELEASE=2.3.1
# Install test dependencies
RUN yum install -y skopeo && \
    export OS=$(go env GOOS) && \
    export ARCH=$(go env GOARCH) && \
    curl -L "https://go.kubebuilder.io/dl/${KUBEBUILDER_RELEASE}/${OS}/${ARCH}" | tar -xz -C /tmp/ && \
    mv /tmp/kubebuilder_${KUBEBUILDER_RELEASE}_${OS}_${ARCH}/ /usr/local/kubebuilder && \
    export PATH=$PATH:/usr/local/kubebuilder/bin && \
    echo "Kubebuilder installation complete!"
