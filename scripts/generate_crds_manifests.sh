#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
export GOFLAGS="-mod=vendor"

# HELM comes from bingo, see Makefile include and .bingo dir for specifics
source .bingo/variables.env

YQ="go run ./vendor/github.com/mikefarah/yq/v3/"
CONTROLLER_GEN="go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen"

ver=${OLM_VERSION:-"0.0.0-dev"}
tmpdir="$(mktemp -p . -d 2>/dev/null || mktemp -d ./tmpdir.XXXXXXX)"
chartdir="${tmpdir}/chart"
crddir="${chartdir}/crds"
crdsrcdir="${tmpdir}/operators"

SED="sed"
if ! command -v ${SED} &> /dev/null; then
    SED="sed"
fi

# OSX distrubtions do not include GNU sed by default, the test below
# exits if we detect that the the insert command is not supported by
# the sed exectuable.
touch "${tmpdir}/sed-test.tmp"
SED_EXIT_CODE=0
${SED} -n -i '1d' "${tmpdir}/sed-test.tmp" &> /dev/null || SED_EXIT_CODE=$?
if [ $SED_EXIT_CODE -ne 0 ]; then
  echo "GNU sed is required for creating manifests, unable to proceed"
  exit $SED_EXIT_CODE
fi

cp -R "${ROOT_DIR}/staging/operator-lifecycle-manager/deploy/chart/" "${chartdir}"
cp "${ROOT_DIR}"/values*.yaml "${tmpdir}"
cp -R "${ROOT_DIR}/staging/api/pkg/operators/" ${crdsrcdir}
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
    mv -v "$f" "${crddir}/0000_50_olm_00-$(basename $f | ${SED} 's/^.*_\([^.]\+\)\.yaml/\1.crd.yaml/')"
done

${SED} -i "s/^[Vv]ersion:.*\$/version: ${ver}/" "${chartdir}/Chart.yaml"

# apply local crc testing patches if necessary
# CRC_E2E_VALUES contains the path to the values file used for running olm on crc locally
# this file is generated in scripts/crc-deploy.sh
CRC_E2E_VALUES=${CRC_E2E_VALUES-""}
CRC_E2E=("")
if ! [ "${CRC_E2E_VALUES}" = "" ]; then
  CRC_E2E=(-f "${tmpdir}/${CRC_E2E_VALUES}")
fi

${HELM} template -n olm -f "${tmpdir}/values.yaml" ${CRC_E2E[@]} --include-crds --output-dir "${chartdir}" "${chartdir}"
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
         ${YQ} w -d'*' --inplace --style=double "$g" 'metadata.annotations['capability.openshift.io/name']' OperatorLifecycleManager
         ${YQ} d -d'*' --inplace "$g" 'spec.template.spec.nodeSelector."node-role.kubernetes.io/master"'
      fi
      ${YQ} w -d'*' --inplace --style=double "$f" 'metadata.annotations['include.release.openshift.io/self-managed-high-availability']' true
      ${YQ} w -d'*' --inplace --style=double "$f" 'metadata.annotations['capability.openshift.io/name']' OperatorLifecycleManager
   done
}

${YQ} merge --inplace -d'*' manifests/0000_50_olm_00-namespace.yaml scripts/namespaces.patch.yaml
${YQ} merge --inplace --arrays=append -d'1' manifests/0000_50_olm_[0-9][0-9]*-olm-operator.serviceaccount.yaml scripts/olm-service.patch.yaml
${YQ} merge --inplace -d'0' manifests/0000_50_olm_00-namespace.yaml scripts/monitoring-namespace.patch.yaml
${YQ} write --inplace -s scripts/olm-deployment.patch.yaml manifests/0000_50_olm_07-olm-operator.deployment.yaml
${YQ} write --inplace -s scripts/catalog-deployment.patch.yaml manifests/0000_50_olm_08-catalog-operator.deployment.yaml
${YQ} write --inplace -s scripts/packageserver-deployment.patch.yaml manifests/0000_50_olm_15-packageserver.clusterserviceversion.yaml
${YQ} merge --inplace manifests/0000_50_olm_[0-9][0-9]*-olmconfig.yaml scripts/cluster-olmconfig.patch.yaml
mv manifests/0000_50_olm_15-packageserver.clusterserviceversion.yaml pkg/manifests/csv.yaml
cp scripts/packageserver-pdb.yaml manifests/0000_50_olm_00-packageserver.pdb.yaml

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
  - name: kube-rbac-proxy
    from:
      kind: DockerImage
      name: quay.io/openshift/origin-kube-rbac-proxy:latest
EOF

cat << EOF > manifests/0000_50_olm_06-psm-operator.deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: package-server-manager
  namespace: openshift-operator-lifecycle-manager
  labels:
    app: package-server-manager
  annotations:
spec:
  strategy:
    type: Recreate
  replicas: 1
  selector:
    matchLabels:
      app: package-server-manager
  template:
    metadata:
      annotations:
        target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
        openshift.io/required-scc: restricted-v2
      labels:
        app: package-server-manager
    spec:
      securityContext:
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
      serviceAccountName: olm-operator-serviceaccount
      priorityClassName: "system-cluster-critical"
      containers:
        - args:
          - --secure-listen-address=0.0.0.0:8443
          - --upstream=http://127.0.0.1:9090/
          - --tls-cert-file=/etc/tls/private/tls.crt
          - --tls-private-key-file=/etc/tls/private/tls.key
          - --logtostderr=true
          image: quay.io/openshift/origin-kube-rbac-proxy:latest
          imagePullPolicy: IfNotPresent
          name: kube-rbac-proxy
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
          ports:
          - containerPort: 8443
            name: metrics
            protocol: TCP
          resources:
            requests:
              memory: 20Mi
              cpu: 10m
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: FallbackToLogsOnError
          volumeMounts:
          - mountPath: /etc/tls/private
            name: package-server-manager-serving-cert
        - name: package-server-manager
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
          command:
            - /bin/psm
            - start
          args:
            - --name
            - \$(PACKAGESERVER_NAME)
            - --namespace
            - \$(PACKAGESERVER_NAMESPACE)
            - "--metrics=:9090"
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
            - name: GOMEMLIMIT
              value: "5MiB"
          resources:
            requests:
              cpu: 10m
              memory: 10Mi
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
      volumes:
      - name: package-server-manager-serving-cert
        secret:
          secretName: package-server-manager-serving-cert
EOF

cat << EOF > manifests/0000_50_olm_06-psm-operator.service.yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    include.release.openshift.io/self-managed-high-availability: "true"
    service.alpha.openshift.io/serving-cert-secret-name: package-server-manager-serving-cert
  name: package-server-manager-metrics
  namespace: openshift-operator-lifecycle-manager
spec:
  ports:
  - name: metrics
    port: 8443
    protocol: TCP
    targetPort: metrics
  selector:
    app: package-server-manager
  sessionAffinity: None
  type: ClusterIP
EOF

cat << EOF > manifests/0000_50_olm_06-psm-operator.servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: package-server-manager-metrics
  namespace: openshift-operator-lifecycle-manager
  annotations:
    include.release.openshift.io/self-managed-high-availability: "true"
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    interval: 30s
    port: metrics
    scheme: https
    tlsConfig:
      caFile: /etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt
      serverName: package-server-manager-metrics.openshift-operator-lifecycle-manager.svc
  namespaceSelector:
    matchNames:
    - openshift-operator-lifecycle-manager
  selector: {}
EOF

cat << EOF > manifests/0000_50_olm_00-pprof-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/hypershift: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
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
    include.release.openshift.io/hypershift: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
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
    include.release.openshift.io/hypershift: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
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
    include.release.openshift.io/hypershift: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
  name: collect-profiles
  namespace: openshift-operator-lifecycle-manager
EOF

cat << EOF > manifests/0000_50_olm_00-pprof-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/hypershift: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    release.openshift.io/create-only: "true"
    openshift.io/owning-component: "Operator Framework / operator-lifecycle-manager"
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
    include.release.openshift.io/hypershift: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
  name: collect-profiles
  namespace: openshift-operator-lifecycle-manager
spec:
  schedule: "*/15 * * * *"
  concurrencyPolicy: "Replace"
  jobTemplate:
    spec:
      template:
        metadata:
          annotations:
            target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
            openshift.io/required-scc: restricted-v2
        spec:
          securityContext:
            runAsNonRoot: true
            seccompProfile:
              type: RuntimeDefault
          serviceAccountName: collect-profiles
          priorityClassName: openshift-user-critical
          containers:
          - name: collect-profiles
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop: ["ALL"]
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
            terminationMessagePolicy: FallbackToLogsOnError
          volumes:
          - name: config-volume
            configMap:
              name: collect-profiles-config
          - name: secret-volume
            secret:
              secretName: pprof-cert
          restartPolicy: Never
EOF

cat << EOF > manifests/0000_50_olm_11-olm-operators.configmap.removed.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: olm-operators
  namespace: openshift-operator-lifecycle-manager
  annotations:
    release.openshift.io/delete: "true"
EOF

cat << EOF > manifests/0000_50_olm_12-olm-operators.catalogsource.removed.yaml
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: olm-operators
  namespace: openshift-operator-lifecycle-manager
  annotations:
    release.openshift.io/delete: "true"

EOF

cat << EOF > manifests/0000_50_olm_14-packageserver.subscription.removed.yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: packageserver
  namespace: openshift-operator-lifecycle-manager
  annotations:
    release.openshift.io/delete: "true"
EOF

cat << EOF > manifests/0000_50_olm_15-csv-viewer.rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  name: copied-csv-viewer
  namespace: openshift
rules:
- apiGroups:
  - "operators.coreos.com"
  resources:
  - "clusterserviceversions"
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  name: copied-csv-viewers
  namespace: openshift
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: copied-csv-viewer
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: system:authenticated
EOF

add_ibm_managed_cloud_annotations "${ROOT_DIR}/manifests"

hypershift_manifests_dir="${ROOT_DIR}/manifests"
for f in $(find "${hypershift_manifests_dir}" -type f -name "*.yaml"); do
   if [[ ! "$f" =~ ".deployment.yaml" ]] && [[ ! "$(basename "${f}")" =~ "psm-operator" ]]; then
      ${YQ} w -d'*' --inplace --style=double "$f" 'metadata.annotations['include.release.openshift.io/hypershift']' true
   fi
done

find "${ROOT_DIR}/manifests" -type f -exec $SED -i "/^#/d" {} \;
find "${ROOT_DIR}/manifests" -type f -exec $SED -i "1{/---/d}" {} \;

# (anik120): uncomment this once https://issues.redhat.com/browse/OLM-2695 is Done.
#${YQ} delete --inplace -d'1' manifests/0000_50_olm_00-namespace.yaml 'metadata.labels."pod-security.kubernetes.io/enforce*"'

# Unlike the namespaces shipped in the upstream version, the openshift-operator-lifecycle-manager and openshift-operator
# namespaces enforce restricted PSA policies, so warnings and audits labels are not neccessary.
${YQ} delete --inplace -d'*' manifests/0000_50_olm_00-namespace.yaml 'metadata.labels."pod-security.kubernetes.io/warn*"'
${YQ} delete --inplace -d'*' manifests/0000_50_olm_00-namespace.yaml 'metadata.labels."pod-security.kubernetes.io/audit*"'

# Let's copy all the manifests to a separate directory for microshift
mkdir -p "${ROOT_DIR}/microshift-manifests/"
rm -rf "${ROOT_DIR}/microshift-manifests/"*
cp "${ROOT_DIR}"/manifests/* "${ROOT_DIR}/microshift-manifests/"

# Let's generate the microshift-manifests.
# There are some differences that we need to take care of:
# - The manifests require a kustomization.yaml file
# - We don't need the specific ibm-cloud-managed manifests
# - We need to adapt some of the manifests to be compatible with microshift as there's no
#   ClusterVersion or ClusterOperator in microshift

# Create the kustomization file
cat << EOF > "${ROOT_DIR}/microshift-manifests/kustomization.yaml"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
EOF

# Now let's generate the kustomization.yaml file for microshift to pick up the manifests.
microshift_manifests_files=$(find "${ROOT_DIR}/microshift-manifests" -type f -name "*.yaml")
# Let's sort the files so that we can have a deterministic order
microshift_manifests_files=$(echo "${microshift_manifests_files}" | sort)
# files to ignore, substring match.
files_to_ignore=("ibm-cloud-managed.yaml" "kustomization.yaml" "psm-operator" "removed" "collect-profiles" "prometheus" "service-monitor" "operatorstatus")

# Add all the manifests files to the kustomization file while ignoring the files in the files_to_ignore list
for file in ${microshift_manifests_files}; do
  for file_to_ignore in ${files_to_ignore[@]}; do
    if [[ "${file}" =~ "${file_to_ignore}" ]]; then
     continue 2
    fi
  done
  echo "  - $(realpath --relative-to "${ROOT_DIR}/microshift-manifests" "${file}")" >> "${ROOT_DIR}/microshift-manifests/kustomization.yaml"
done

# Now we need to get rid of these args from the olm-operator deployment:
#
# - --writeStatusName
# - operator-lifecycle-manager
# - --writePackageServerStatusName
# - operator-lifecycle-manager-packageserver
#
${SED} -i '/- --writeStatusName/,+3d' ${ROOT_DIR}/microshift-manifests/0000_50_olm_07-olm-operator.deployment.yaml

# Replace the namespace openshift, as it doesn't exist on microshift, in the rbac file
${SED} -i 's/  namespace: openshift/  namespace: openshift-operator-lifecycle-manager/g' ${ROOT_DIR}/microshift-manifests/0000_50_olm_15-csv-viewer.rbac.yaml
