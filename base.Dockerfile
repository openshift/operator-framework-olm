# Dockerfile to bootstrap build and test in openshift-ci

FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.17-openshift-4.11

ARG KUBEBUILDER_RELEASE=2.3.1
# Install test dependencies
# TODO(tflannag): This is a quick fix to kubebuilder's quick start instructions
# that were failing e2e tests. We'll want to update our o/release ci-operator
# configuration to allow build_root changes to be reflected in a PR vs.
# instead of the CI pipeline always defaulting to building the HEAD version
# of the dockerfile:
# - https://docs.ci.openshift.org/docs/architecture/ci-operator/#build-root-image
# Note(tflannag): We ran into some issues curling from the https://go.kubebuilder.io/dl
# domain as the output file was HTLM-based, so curl from the github releases
# until this has been resolved.
RUN yum install -y skopeo && \
    export OS=$(go env GOOS) && \
    export ARCH=$(go env GOARCH) && \
    curl -L "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_RELEASE}/kubebuilder_${KUBEBUILDER_RELEASE}_${OS}_${ARCH}.tar.gz" | tar -xz -C /tmp/ && \
    mv /tmp/kubebuilder_${KUBEBUILDER_RELEASE}_${OS}_${ARCH}/ /usr/local/kubebuilder && \
    export PATH=$PATH:/usr/local/kubebuilder/bin && \
    kubebuilder version && \
    echo "Kubebuilder installation complete!"
