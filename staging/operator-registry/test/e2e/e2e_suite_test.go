package e2e_test

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dockerUsername = os.Getenv("DOCKER_USERNAME")
	dockerPassword = os.Getenv("DOCKER_PASSWORD")
	dockerHost     = os.Getenv("DOCKER_REGISTRY_HOST") // 'DOCKER_HOST' is reserved for the docker daemon
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var _ = BeforeSuite(func() {
	switch {
	case dockerUsername == "" && dockerPassword == "" && dockerHost == "":
		// No registry credentials or local registry host provided
		// Fail early
		GinkgoT().Fatal("No registry credentials or local registry host provided")
	case dockerHost != "" && dockerUsername == "" && dockerPassword == "":
		// Running against local secure registry without credentials
		// No need to login
		return
	}

	// FIXME: Since podman login doesn't work with daemonless image pulling, we need to login with docker first so podman tests don't fail.
	dockerlogin := exec.Command("docker", "login", "-u", dockerUsername, "-p", dockerPassword, dockerHost)
	out, err := dockerlogin.CombinedOutput()
	Expect(err).ToNot(HaveOccurred(), "Error logging into %s: %s", dockerHost, out)

	By(fmt.Sprintf("Using container image registry %s", dockerHost))
})
