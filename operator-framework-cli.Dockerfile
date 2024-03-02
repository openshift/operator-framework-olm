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

# copy and build vendored grpc_health_probe
RUN CGO_ENABLED=0 go build -mod=vendor -tags netgo -ldflags "-w" ./vendor/github.com/grpc-ecosystem/grpc-health-probe/...

FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.21-openshift-4.16 AS staticbuilder
# Permit opm binary to be compiled statically. The Red Hat compiler
# provided by ART will otherwise force FIPS compliant dynamic compilation.
ENV GO_COMPLIANCE_EXCLUDE="build.*operator-registry/cmd/opm"
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 make build/registry
FROM registry.ci.openshift.org/ocp/4.16:base-rhel9

COPY --from=builder /src/bin/* /bin/registry/
COPY --from=builder /src/grpc-health-probe /bin/grpc_health_probe
# copy the dynamically-linked versions to /bin/registry with a -rhel8 suffix
COPY --from=builder-rhel8 /src/bin/opm /bin/registry/opm-rhel8
COPY --from=builder-rhel8 /src/bin/registry-server /bin/registry/registry-server-rhel8
COPY --from=builder-rhel8 /src/bin/initializer /bin/registry/initializer-rhel8
COPY --from=builder-rhel8 /src/bin/configmap-server /bin/registry/configmap-server-rhel8

COPY --from=staticbuilder /src/bin/opm /bin/registry/opm-static

RUN ln -s /bin/registry/* /bin

RUN mkdir /registry
RUN chgrp -R 0 /registry && \
    chmod -R g+rwx /registry
WORKDIR /registry

USER 1001
EXPOSE 50051

ENTRYPOINT ["/bin/registry-server"]
CMD ["--database", "/bundles.db"]

LABEL io.k8s.display-name="OpenShift Operator Registry" \
      io.k8s.description="This is a component of OpenShift Operator Lifecycle Manager and is the base for operator catalog API containers." \
      maintainer="Odin Team <aos-odin@redhat.com>" \
      summary="Operator Registry runs in a Kubernetes or OpenShift cluster to provide operator catalog data to Operator Lifecycle Manager."
