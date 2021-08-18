#! /bin/bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
export GOFLAGS="-mod=vendor"

YQ="go run ./vendor/github.com/mikefarah/yq/v3/"
CONTROLLER_GEN="go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen"
HELM="go run helm.sh/helm/v3/cmd/helm"

ver=$(cat ./staging/operator-lifecycle-manager/OLM_VERSION)
tmpdir="$(mktemp -p . -d 2>/dev/null || mktemp -d ./tmpdir.XXXXXXX)"
chartdir="${tmpdir}/chart"
crddir="${chartdir}/crds"
crdsrcdir="${tmpdir}/operators"

cp -R "${ROOT_DIR}/staging/operator-lifecycle-manager/deploy/chart/" "${chartdir}"
cp "${ROOT_DIR}/values.yaml" "${tmpdir}"
ln -snf $(realpath --relative-to ${tmpdir} ${ROOT_DIR}/staging/api/pkg/operators/) ${crdsrcdir}
rm -rf ./manifests/* ${crddir}/*

trap "rm -rf ${tmpdir}" EXIT

${CONTROLLER_GEN} crd:crdVersions=v1 output:crd:dir=${crddir} paths=${crdsrcdir}/...
${CONTROLLER_GEN} schemapatch:manifests=${crddir} output:dir=${crddir} paths=${crdsrcdir}/...

${YQ} w --inplace ${crddir}/operators.coreos.com_clusterserviceversions.yaml spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.spec.properties.containers.items.properties.ports.items.properties.protocol.default TCP
${YQ} w --inplace ${crddir}/operators.coreos.com_clusterserviceversions.yaml spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.spec.properties.initContainers.items.properties.ports.items.properties.protocol.default TCP
${YQ} w --inplace ${crddir}/operators.coreos.com_clusterserviceversions.yaml spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.install.properties.spec.properties.deployments.items.properties.spec.properties.template.properties.metadata.x-kubernetes-preserve-unknown-fields true
${YQ} d --inplace ${crddir}/operators.coreos.com_operatorconditions.yaml 'spec.versions[*].schema.openAPIV3Schema.properties.spec.properties.overrides.items.required(.==lastTransitionTime)'

for f in ${crddir}/*.yaml ; do
    ${YQ} d --inplace $f status
    mv -v "$f" "${crddir}/0000_50_olm_00-$(basename $f | sed 's/^.*_\([^.]\+\)\.yaml/\1.crd.yaml/')"
done

sed -i "s/^[Vv]ersion:.*\$/version: ${ver}/" "${chartdir}/Chart.yaml"

${HELM} template -n olm -f "${tmpdir}/values.yaml" --include-crds --output-dir "${chartdir}" "${chartdir}"
cp -R "${chartdir}"/olm/{templates,crds}/. "./manifests"

add_ibm_managed_cloud_annotations() {
   local manifests_dir=$1

   for f in "${manifests_dir}"/*.yaml; do
      if [[ ! "$(basename "${f}")" =~ .*\.deployment\..* ]]; then
         ${YQ} w -d'*' --inplace --style=double "$f" 'metadata.annotations['include.release.openshift.io/ibm-cloud-managed']' true
      else
         g="${f/%.yaml/.ibm-cloud-managed.yaml}"
         cp "${f}" "${g}"
         ${YQ} w -d'*' --inplace --style=double "$g" 'metadata.annotations['include.release.openshift.io/ibm-cloud-managed']' true
         ${YQ} d -d'*' --inplace "$g" 'spec.template.spec.nodeSelector."node-role.kubernetes.io/master"'
      fi
      ${YQ} w -d'*' --inplace --style=double "$f" 'metadata.annotations['include.release.openshift.io/self-managed-high-availability']' true
      ${YQ} w -d'*' --inplace --style=double "$f" 'metadata.annotations['include.release.openshift.io/single-node-developer']' true
   done
}

${YQ} merge --inplace -d'*' manifests/0000_50_olm_00-namespace.yaml scripts/namespaces.patch.yaml
${YQ} merge --inplace -d'0' manifests/0000_50_olm_00-namespace.yaml scripts/monitoring-namespace.patch.yaml
${YQ} write --inplace -s scripts/olm-deployment.patch.yaml manifests/0000_50_olm_07-olm-operator.deployment.yaml
${YQ} write --inplace -s scripts/catalog-deployment.patch.yaml manifests/0000_50_olm_08-catalog-operator.deployment.yaml
${YQ} write --inplace -s scripts/packageserver-deployment.patch.yaml manifests/0000_50_olm_15-packageserver.clusterserviceversion.yaml
mv manifests/0000_50_olm_15-packageserver.clusterserviceversion.yaml pkg/manifests/csv.yaml

cat << EOF > manifests/image-references
kind: ImageStream
apiVersion: image.openshift.io/v1
spec:
  tags:
  - name: operator-lifecycle-manager
    from:
      kind: DockerImage
      name: "$(${YQ} read values.yaml olm.image.ref)"
  - name: operator-registry
    from:
      kind: DockerImage
      name: quay.io/operator-framework/configmap-operator-registry:latest
EOF

cat << EOF > manifests/0000_50_olm-06-psm-operator.deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: package-server-manager
  namespace: openshift-operator-lifecycle-manager
  labels:
    app: package-server-manager
  annotations:
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
spec:
  strategy:
    type: RollingUpdate
  replicas: 1
  selector:
    matchLabels:
      app: package-server-manager
  template:
    metadata:
      annotations:
        target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
      labels:
        app: package-server-manager
    spec:
      serviceAccountName: olm-operator-serviceaccount
      priorityClassName: "system-cluster-critical"
      containers:
        - name: package-server-manager
          command:
            - /bin/psm
            - start
          args:
            - --name
            - \$(PACKAGESERVER_NAME)
            - --namespace
            - \$(PACKAGESERVER_NAMESPACE)
          image: quay.io/operator-framework/olm@sha256:de396b540b82219812061d0d753440d5655250c621c753ed1dc67d6154741607
          imagePullPolicy: IfNotPresent
          env:
            - name: PACKAGESERVER_NAME
              value: packageserver
            - name: PACKAGESERVER_IMAGE
              value: quay.io/operator-framework/olm@sha256:de396b540b82219812061d0d753440d5655250c621c753ed1dc67d6154741607
            - name: PACKAGESERVER_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: RELEASE_VERSION
              value: "0.0.1-snapshot"
          resources:
            requests:
              cpu: 10m
              memory: 50Mi
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 30
          readinessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 30
          terminationMessagePolicy: FallbackToLogsOnError
      nodeSelector:
        kubernetes.io/os: linux
        node-role.kubernetes.io/master: ""
      tolerations:
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
          operator: Exists
        - effect: NoExecute
          key: node.kubernetes.io/unreachable
          operator: Exists
          tolerationSeconds: 120
        - effect: NoExecute
          key: node.kubernetes.io/not-ready
          operator: Exists
          tolerationSeconds: 120
EOF

cat << EOF > manifests/0000_50_olm_00-pprof-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
    release.openshift.io/create-only: "true"
  name: collect-profiles-config
  namespace: openshift-operator-lifecycle-manager
data:
  pprof-config.yaml: |
    disabled: False
EOF

cat << EOF > manifests/0000_50_olm_00-pprof-rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: collect-profiles
  namespace: openshift-operator-lifecycle-manager
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "create", "delete"]
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: collect-profiles
  namespace: openshift-operator-lifecycle-manager
subjects:
- kind: ServiceAccount
  name: collect-profiles
  namespace: openshift-operator-lifecycle-manager
roleRef:
  kind: Role
  name: collect-profiles
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: collect-profiles
  namespace: openshift-operator-lifecycle-manager
EOF

cat << EOF > manifests/0000_50_olm_00-pprof-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
    release.openshift.io/create-only: "true"
  name: pprof-cert
  namespace: openshift-operator-lifecycle-manager
type: kubernetes.io/tls
data:
  tls.crt: ""
  tls.key: ""
EOF

cat << EOF > manifests/0000_50_olm_07-collect-profiles.cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: collect-profiles
  namespace: openshift-operator-lifecycle-manager
spec:
  schedule: "*/15 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: collect-profiles
          priorityClassName: openshift-user-critical
          containers:
          - name: collect-profiles
            image: quay.io/operator-framework/olm@sha256:de396b540b82219812061d0d753440d5655250c621c753ed1dc67d6154741607
            imagePullPolicy: IfNotPresent
            command:
            - bin/collect-profiles
            args: 
            - -n
            - openshift-operator-lifecycle-manager
            - --config-mount-path
            - /etc/config
            - --cert-mount-path
            - /var/run/secrets/serving-cert
            - olm-operator-heap-:https://olm-operator-metrics:8443/debug/pprof/heap
            - catalog-operator-heap-:https://catalog-operator-metrics:8443/debug/pprof/heap
            volumeMounts:
            - mountPath: /etc/config
              name: config-volume
            - mountPath: /var/run/secrets/serving-cert
              name: secret-volume
            resources:
              requests:
                cpu: 10m
                memory: 80Mi
          volumes:
          - name: config-volume
            configMap:
              name: collect-profiles-config
          - name: secret-volume
            secret:
              secretName: pprof-cert
          restartPolicy: Never
EOF

add_ibm_managed_cloud_annotations "${ROOT_DIR}/manifests"

# requires gnu sed if on mac
find "${ROOT_DIR}/manifests" -type f -exec sed -i "/^#/d" {} \;
find "${ROOT_DIR}/manifests" -type f -exec sed -i "1{/---/d}" {} \;
