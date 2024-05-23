FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.22-openshift-4.17 AS builder

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /src

COPY . .
RUN make build/registry cross

# copy and build vendored grpc_health_probe
RUN CGO_ENABLED=0 go build -mod=vendor -tags netgo -ldflags "-w" ./vendor/github.com/grpc-ecosystem/grpc-health-probe/...

FROM registry.ci.openshift.org/ocp/4.17:base-rhel9

COPY --from=builder /src/bin/* /bin/registry/
COPY --from=builder /src/grpc-health-probe /bin/grpc_health_probe

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
