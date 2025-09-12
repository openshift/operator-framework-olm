package util

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	o "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	testdata "github.com/openshift/operator-framework-olm/tests-extension/pkg/bindata/qe"
)

func init() {
	if KubeConfigPath() == "" {
		fmt.Fprintf(os.Stderr, "Please set KUBECONFIG first!\n")
		os.Exit(0)
	}
}

// WaitForServiceAccount waits until the named service account gets fully
// provisioned
func WaitForServiceAccount(c corev1client.ServiceAccountInterface, name string, checkSecret bool) error {
	countOutput := -1
	// add Logf for better debug, but it will possible generate many logs because of 100 millisecond
	// so, add countOutput so that it output log every 100 times (10s)
	waitFn := func(ctx context.Context) (bool, error) {
		countOutput++
		sc, err := c.Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			// If we can't access the service accounts, let's wait till the controller
			// create it.
			if errors.IsNotFound(err) || errors.IsForbidden(err) {
				if countOutput%100 == 0 {
					e2e.Logf("Waiting for service account %q to be available: %v (will retry) ...", name, err)
				}
				return false, nil
			}
			return false, fmt.Errorf("failed to get service account %q: %v", name, err)
		}
		secretNames := []string{}
		var hasDockercfg bool
		for _, s := range sc.Secrets {
			if strings.Contains(s.Name, "dockercfg") {
				hasDockercfg = true
			}
			secretNames = append(secretNames, s.Name)
		}
		if hasDockercfg || !checkSecret {
			return true, nil
		}
		if countOutput%100 == 0 {
			e2e.Logf("Waiting for service account %q secrets (%s) to include dockercfg ...", name, strings.Join(secretNames, ","))
		}
		return false, nil
	}
	return wait.PollUntilContextTimeout(context.TODO(), 100*time.Millisecond, 3*time.Minute, false, waitFn)
}

// GetPodNamesByFilter looks up pods that satisfy the predicate and returns their names.
func GetPodNamesByFilter(c corev1client.PodInterface, label labels.Selector, predicate func(corev1.Pod) bool) ([]string, error) {
	podList, err := c.List(context.Background(), metav1.ListOptions{LabelSelector: label.String()})
	if err != nil {
		return nil, err
	}
	var podNames []string
	for _, pod := range podList.Items {
		if predicate(pod) {
			podNames = append(podNames, pod.Name)
		}
	}
	return podNames, nil
}

// WaitForPods waits until given number of pods that match the label selector and
// satisfy the predicate are found
func WaitForPods(c corev1client.PodInterface, label labels.Selector, predicate func(corev1.Pod) bool, count int, timeout time.Duration) ([]string, error) {
	var podNames []string
	err := wait.PollUntilContextTimeout(context.TODO(), 1*time.Second, timeout, false, func(ctx context.Context) (bool, error) {
		p, e := GetPodNamesByFilter(c, label, predicate)
		if e != nil {
			return true, e
		}
		if len(p) != count {
			return false, nil
		}
		podNames = p
		return true, nil
	})
	return podNames, err
}

// CheckPodIsRunning returns true if the pod is running
func CheckPodIsRunning(pod corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning
}

// CheckPodIsSucceeded returns true if the pod status is "Succdeded"
func CheckPodIsSucceeded(pod corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodSucceeded
}

// CheckPodIsReady returns true if the pod's ready probe determined that the pod is ready.
func CheckPodIsReady(pod corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}
	for _, cond := range pod.Status.Conditions {
		if cond.Type != corev1.PodReady {
			continue
		}
		return cond.Status == corev1.ConditionTrue
	}
	return false
}

// KubeConfigPath returns the value of KUBECONFIG environment variable
func KubeConfigPath() string {
	// can't use gomega in this method since it is used outside of It()
	return os.Getenv("KUBECONFIG")
}

// ArtifactDirPath returns the value of ARTIFACT_DIR environment variable
func ArtifactDirPath() string {
	path := os.Getenv("ARTIFACT_DIR")
	o.Expect(path).NotTo(o.BeNil())
	o.Expect(path).NotTo(o.BeEmpty())
	return path
}

// ArtifactPath returns the absolute path to the fix artifact file
// The path is relative to ARTIFACT_DIR
func ArtifactPath(elem ...string) string {
	return filepath.Join(append([]string{ArtifactDirPath()}, elem...)...)
}

var (
	fixtureDirLock sync.Once
	fixtureDir     string
)

// FixturePath returns an absolute path to a fixture file in test/qe/testdata/,
// test/integration/, or examples/.
func FixturePath(elem ...string) string {
	switch {
	case len(elem) == 0:
		panic("must specify path")
	case len(elem) > 3 && elem[0] == ".." && elem[1] == ".." && elem[2] == "examples":
		elem = elem[2:]
	case len(elem) > 3 && elem[0] == ".." && elem[1] == ".." && elem[2] == "install":
		elem = elem[2:]
	case len(elem) > 3 && elem[0] == ".." && elem[1] == "integration":
		elem = append([]string{"test"}, elem[1:]...)
	case elem[0] == "testdata":
		elem = append([]string{"test", "qe"}, elem...)
	default:
		panic(fmt.Sprintf("Fixtures must be in test/qe/testdata or examples not %s", path.Join(elem...)))
	}
	fixtureDirLock.Do(func() {
		dir, err := os.MkdirTemp("", "fixture-testdata-dir")
		if err != nil {
			panic(err)
		}
		fixtureDir = dir
	})
	relativePath := path.Join(elem...)
	fullPath := path.Join(fixtureDir, relativePath)
	if err := testdata.RestoreAsset(fixtureDir, relativePath); err != nil {
		if err := testdata.RestoreAssets(fixtureDir, relativePath); err != nil {
			panic(err)
		}
		if err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err := os.Chmod(path, 0640); err != nil {
				return err
			}
			if stat, err := os.Lstat(path); err == nil && stat.IsDir() {
				return os.Chmod(path, 0755)
			}
			return nil
		}); err != nil {
			panic(err)
		}
	} else {
		if err := os.Chmod(fullPath, 0640); err != nil {
			panic(err)
		}
	}

	p, err := filepath.Abs(fullPath)
	if err != nil {
		panic(err)
	}
	return p
}

// ParseLabelsOrDie turns the given string into a label selector or
// panics; for tests or other cases where you know the string is valid.
// TODO: Move this to the upstream labels package.
func ParseLabelsOrDie(str string) labels.Selector {
	ret, err := labels.Parse(str)
	if err != nil {
		panic(fmt.Sprintf("cannot parse '%v': %v", str, err))
	}
	return ret
}
