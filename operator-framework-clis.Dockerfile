FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.21-openshift-4.16 AS builder-rhel8
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
WORKDIR /src
COPY . .
RUN make build/registry cross

FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.21-openshift-4.16 AS builder
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
WORKDIR /src
COPY . .
RUN make build/registry cross

FROM scratch 

COPY --from=builder /src/bin/opm /clis/opm-rhel9
COPY --from=builder /src/bin/darwin-amd64-opm /clis/darwin-amd64-opm
COPY --from=builder /src/bin/windows-amd64-opm /clis/windows-amd64-opm

# copy the dynamically-linked versions to /clis with a -rhel8 suffix
COPY --from=builder-rhel8 /src/bin/opm /clis/opm-rhel8

USER 1001

LABEL io.k8s.display-name="OpenShift Operator Framework CLIs" \
      io.k8s.description="This is a non-runnable image containing binary builds of various Operator Framework CLI tools, primarily used to publish binaries to the OpenShift mirror." \
      maintainer="Odin Team <aos-odin@redhat.com>"
