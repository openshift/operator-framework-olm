FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.20-openshift-4.14 as builder

WORKDIR /src
COPY main.go go.mod ./
RUN go build -o /bin/bumper -mod=mod ./...

FROM quay.io/centos/centos:stream8

RUN dnf install -y git glibc make
COPY --from=builder /bin/bumper /usr/bin/bumper
COPY --from=builder /usr/bin/go /usr/bin/go
COPY --from=builder /usr/lib/golang /usr/lib/golang

ENTRYPOINT ["bumper"]