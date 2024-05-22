FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.22-openshift-4.17 AS builder-rhel8
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
WORKDIR /src
COPY . .
RUN make build/registry cross

FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.22-openshift-4.17 AS builder
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
WORKDIR /src
COPY . .
RUN make build/registry cross

FROM scratch 

COPY --from=builder /src/bin/opm /tools/opm-rhel9
COPY --from=builder /src/bin/darwin-amd64-opm /tools/darwin-amd64-opm
COPY --from=builder /src/bin/windows-amd64-opm /tools/windows-amd64-opm

# copy the dynamically-linked versions to /tools with a -rhel8 suffix
COPY --from=builder-rhel8 /src/bin/opm /tools/opm-rhel8

USER 1001

LABEL io.k8s.display-name="OpenShift Operator Framework Tools" \
      io.k8s.description="This is a non-runnable image containing binary builds of various Operator Framework tools, primarily used to publish binaries to the OpenShift mirror." \
      maintainer="Odin Team <aos-odin@redhat.com>"
