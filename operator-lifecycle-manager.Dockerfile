FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.22-openshift-4.17 AS builder

ENV GO111MODULE auto
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Permit the cpb and copy-content binaries to be compiled statically. The Red Hat compiler
# provided by ART will otherwise force FIPS compliant dynamic compilation.
ENV GO_COMPLIANCE_EXCLUDE="build.*operator-lifecycle-manager/(util/cpb|cmd/copy-content)"

WORKDIR /build

# copy just enough of the git repo to parse HEAD, used to record version in OLM binaries
COPY .git/HEAD .git/HEAD
COPY .git/refs/heads/. .git/refs/heads
RUN mkdir -p .git/objects

COPY . .
RUN make build/olm bin/cpb

FROM registry.ci.openshift.org/ocp/4.17:base-rhel9

ADD manifests/ /manifests
LABEL io.openshift.release.operator=true

# Copy the binary to a standard location where it will run.
COPY --from=builder /build/bin/olm /bin/olm
COPY --from=builder /build/bin/catalog /bin/catalog
COPY --from=builder /build/bin/collect-profiles /bin/collect-profiles
COPY --from=builder /build/bin/package-server /bin/package-server
COPY --from=builder /build/bin/cpb /bin/cpb
COPY --from=builder /build/bin/psm /bin/psm
COPY --from=builder /build/bin/copy-content /bin/copy-content

# This image doesn't need to run as root user.
USER 1001

EXPOSE 8080
EXPOSE 5443

# Apply labels as needed. ART build automation fills in others required for
# shipping, including component NVR (name-version-release) and image name. OSBS
# applies others at build time. So most required labels need not be in the source.
#
# io.k8s.display-name is required and is displayed in certain places in the
# console (someone correct this if that's no longer the case)
#
# io.k8s.description is equivalent to "description" and should be defined per
# image; otherwise the parent image's description is inherited which is
# confusing at best when examining images.
#
LABEL io.k8s.display-name="OpenShift Operator Lifecycle Manager" \
      io.k8s.description="This is a component of OpenShift Container Platform and manages the lifecycle of operators." \
      maintainer="Odin Team <aos-odin@redhat.com>"
