package olmv0util

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// customSchemaCodec is a gRPC codec that handles raw bytes for request encoding
// and structpb.Struct for response decoding.
type customSchemaCodec struct{}

type rawRequest struct {
	bytes []byte
}

func (customSchemaCodec) Marshal(v any) ([]byte, error) {
	if raw, ok := v.(*rawRequest); ok {
		return raw.bytes, nil
	}
	return nil, fmt.Errorf("customSchemaCodec: unsupported marshal type %T", v)
}

func (customSchemaCodec) Unmarshal(data []byte, v any) error {
	if s, ok := v.(*structpb.Struct); ok {
		return proto.Unmarshal(data, s)
	}
	return fmt.Errorf("customSchemaCodec: unsupported unmarshal type %T", v)
}

func (customSchemaCodec) Name() string { return "proto" }

var _ encoding.Codec = customSchemaCodec{}

// encodeCustomSchemaRequest builds the protobuf wire encoding for
// ExperimentalListPackageCustomSchemasRequest{schema, packageName}.
func encodeCustomSchemaRequest(schema, packageName string) []byte {
	var b []byte
	if schema != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, schema)
	}
	if packageName != "" {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, packageName)
	}
	return b
}

// GetOPMBaseImage extracts the opm image reference from the catalog-operator
// deployment in openshift-operator-lifecycle-manager namespace.
func GetOPMBaseImage(oc *exutil.CLI) string {
	output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
		"deployment", "catalog-operator",
		"-n", "openshift-operator-lifecycle-manager",
		"-o=jsonpath={.spec.template.spec.containers[0].args}",
	).Output()
	o.Expect(err).NotTo(o.HaveOccurred())

	// The output is a JSON array like ["--namespace","olm","--opmImage=registry..."]
	var args []string
	if err := json.Unmarshal([]byte(output), &args); err != nil {
		e2e.Logf("failed to parse args as JSON array, falling back to string split: %v", err)
		args = strings.Fields(strings.NewReplacer("[", "", "]", "", "\"", "", ",", " ").Replace(output))
	}
	for _, arg := range args {
		if strings.HasPrefix(arg, "--opmImage=") {
			image := strings.TrimPrefix(arg, "--opmImage=")
			e2e.Logf("found opm base image from args: %s", image)
			return image
		}
	}

	// Fallback: check container env vars for RELATED_IMAGE_OPM
	envOutput, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
		"deployment", "catalog-operator",
		"-n", "openshift-operator-lifecycle-manager",
		"-o=jsonpath={.spec.template.spec.containers[0].env[?(@.name==\"RELATED_IMAGE_OPM\")].value}",
	).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	if envOutput != "" {
		e2e.Logf("found opm base image from env: %s", envOutput)
		return envOutput
	}

	// Fallback: use the container image itself
	imgOutput, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
		"deployment", "catalog-operator",
		"-n", "openshift-operator-lifecycle-manager",
		"-o=jsonpath={.spec.template.spec.containers[0].image}",
	).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	e2e.Logf("fallback to catalog-operator container image as opm base: %s", imgOutput)
	return imgOutput
}

// BuildCustomCatalogImage builds a catalog image in-cluster using OpenShift
// BuildConfig. It creates an ImageStream and BuildConfig, starts a binary build
// from the provided FBC content, and waits for completion.
// Returns the internal registry image reference.
func BuildCustomCatalogImage(oc *exutil.CLI, namespace, name, baseImage string, fbcContent []byte) string {
	// Create ImageStream
	isTemplate := exutil.FixturePath("testdata", "olm", "custom-schema-imagestream.yaml")
	err := ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", isTemplate,
		"-p", "NAME="+name, "NAMESPACE="+namespace)
	o.Expect(err).NotTo(o.HaveOccurred())
	e2e.Logf("created ImageStream %s/%s", namespace, name)

	// Create BuildConfig
	bcTemplate := exutil.FixturePath("testdata", "olm", "custom-schema-buildconfig.yaml")
	err = ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", bcTemplate,
		"-p", "NAME="+name, "NAMESPACE="+namespace, "BASE_IMAGE="+baseImage)
	o.Expect(err).NotTo(o.HaveOccurred())
	e2e.Logf("created BuildConfig %s/%s with base image %s", namespace, name, baseImage)

	// Prepare build directory
	buildDir, err := os.MkdirTemp("", "custom-schema-build-")
	o.Expect(err).NotTo(o.HaveOccurred())
	defer func() { _ = os.RemoveAll(buildDir) }()

	configsDir := filepath.Join(buildDir, "configs")
	err = os.MkdirAll(configsDir, 0755)
	o.Expect(err).NotTo(o.HaveOccurred())

	err = os.WriteFile(filepath.Join(configsDir, "index.json"), fbcContent, 0644)
	o.Expect(err).NotTo(o.HaveOccurred())

	// Write Dockerfile following the pattern from
	// https://github.com/operator-framework/operator-registry/blob/master/opm-example.Dockerfile
	// The RUN step pre-populates the serve cache so the integrity check passes at runtime.
	dockerfile := "FROM " + baseImage + "\n" +
		"ENTRYPOINT [\"/bin/opm\"]\n" +
		"CMD [\"serve\", \"/configs\", \"--cache-dir=/tmp/cache\"]\n" +
		"COPY configs /configs\n" +
		"RUN [\"/bin/opm\", \"serve\", \"/configs\", \"--cache-dir=/tmp/cache\", \"--cache-only\"]\n" +
		"LABEL operators.operatorframework.io.index.configs.v1=/configs\n"
	err = os.WriteFile(filepath.Join(buildDir, "Dockerfile"), []byte(dockerfile), 0644)
	o.Expect(err).NotTo(o.HaveOccurred())

	// Start binary build
	output, err := oc.AsAdmin().WithoutNamespace().Run("start-build").Args(
		name, "-n", namespace, "--from-dir="+buildDir, "--follow",
	).Output()
	if err != nil {
		e2e.Logf("build output: %s", output)
	}
	o.Expect(err).NotTo(o.HaveOccurred())
	e2e.Logf("build completed for %s/%s", namespace, name)

	// Wait for build to complete successfully
	err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 300*time.Second, false, func(ctx context.Context) (bool, error) {
		phase, getErr := oc.AsAdmin().WithoutNamespace().Run("get").Args(
			"build", name+"-1", "-n", namespace,
			"-o=jsonpath={.status.phase}",
		).Output()
		if getErr != nil {
			e2e.Logf("error checking build status: %v", getErr)
			return false, nil
		}
		if phase == "Complete" {
			return true, nil
		}
		if phase == "Failed" || phase == "Error" || phase == "Cancelled" {
			return false, fmt.Errorf("build %s-1 failed with phase %s", name, phase)
		}
		e2e.Logf("build %s-1 phase: %s", name, phase)
		return false, nil
	})
	o.Expect(err).NotTo(o.HaveOccurred())

	imageRef := fmt.Sprintf("image-registry.openshift-image-registry.svc:5000/%s/%s:latest", namespace, name)
	e2e.Logf("catalog image built: %s", imageRef)
	return imageRef
}

// PortForwardToCatalogPod starts an oc port-forward to the gRPC port (50051)
// of the catalog source pod. Returns the local address and a cleanup function.
func PortForwardToCatalogPod(oc *exutil.CLI, namespace, catalogName string) (string, func()) {
	// Find the catalog source pod
	podName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
		"pods", "-n", namespace,
		"-l", "olm.catalogSource="+catalogName,
		"-o=jsonpath={.items[0].metadata.name}",
	).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(podName).NotTo(o.BeEmpty(), "no pod found for catalog source %s", catalogName)
	e2e.Logf("found catalog pod: %s", podName)

	// Find a free local port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	o.Expect(err).NotTo(o.HaveOccurred())
	localPort := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	// Start port-forward
	kubeconfig := exutil.KubeConfigPath()
	cmd := exec.Command("oc", "--kubeconfig="+kubeconfig,
		"port-forward", "-n", namespace, podName,
		fmt.Sprintf("%d:50051", localPort))

	stdout, err := cmd.StdoutPipe()
	o.Expect(err).NotTo(o.HaveOccurred())

	err = cmd.Start()
	o.Expect(err).NotTo(o.HaveOccurred())
	e2e.Logf("started port-forward to %s on local port %d", podName, localPort)

	// Wait for port-forward to be ready by reading stdout
	scanner := bufio.NewScanner(stdout)
	ready := make(chan struct{})
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			e2e.Logf("port-forward: %s", line)
			if strings.Contains(line, "Forwarding from") {
				close(ready)
				return
			}
		}
	}()

	select {
	case <-ready:
		e2e.Logf("port-forward ready on localhost:%d", localPort)
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		o.Expect(fmt.Errorf("port-forward did not become ready in 30s")).NotTo(o.HaveOccurred())
	}

	localAddr := fmt.Sprintf("localhost:%d", localPort)
	cleanup := func() {
		if cmd.Process != nil {
			e2e.Logf("stopping port-forward (pid %d)", cmd.Process.Pid)
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}
	return localAddr, cleanup
}

// ListPackageCustomSchemas calls the ExperimentalListPackageCustomSchemas gRPC
// endpoint with the x-acknowledge-experimental header and returns the results
// as Go maps (one per streamed Struct message).
func ListPackageCustomSchemas(ctx context.Context, grpcAddr, schema, packageName string) ([]map[string]interface{}, error) {
	return listPackageCustomSchemas(ctx, grpcAddr, schema, packageName, true)
}

// ListPackageCustomSchemasWithoutExperimentalHeader calls the
// ExperimentalListPackageCustomSchemas gRPC endpoint WITHOUT the
// x-acknowledge-experimental header. The server should silently return an
// empty stream in this case.
func ListPackageCustomSchemasWithoutExperimentalHeader(ctx context.Context, grpcAddr, schema, packageName string) ([]map[string]interface{}, error) {
	return listPackageCustomSchemas(ctx, grpcAddr, schema, packageName, false)
}

func listPackageCustomSchemas(ctx context.Context, grpcAddr, schema, packageName string, withExperimentalHeader bool) ([]map[string]interface{}, error) {
	conn, err := grpc.NewClient(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", grpcAddr, err)
	}
	defer func() { _ = conn.Close() }()

	if withExperimentalHeader {
		md := metadata.New(map[string]string{
			"x-acknowledge-experimental": "true",
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	desc := &grpc.StreamDesc{
		StreamName:    "ExperimentalListPackageCustomSchemas",
		ServerStreams: true,
		ClientStreams: false,
	}

	stream, err := conn.NewStream(ctx, desc,
		"/api.ExperimentalRegistry/ExperimentalListPackageCustomSchemas",
		grpc.ForceCodec(customSchemaCodec{}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	req := &rawRequest{bytes: encodeCustomSchemaRequest(schema, packageName)}
	if err := stream.SendMsg(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if err := stream.CloseSend(); err != nil {
		return nil, fmt.Errorf("failed to close send: %w", err)
	}

	var results []map[string]interface{}
	for {
		resp := &structpb.Struct{}
		err := stream.RecvMsg(resp)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to receive response: %w", err)
		}
		results = append(results, resp.AsMap())
	}

	return results, nil
}
