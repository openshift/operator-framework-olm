package e2e

import (
	"context"
	"database/sql"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/operator-framework/operator-registry/pkg/containertools"
	"github.com/operator-framework/operator-registry/pkg/lib/bundle"
	"github.com/operator-framework/operator-registry/pkg/lib/indexer"
	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/operator-framework/operator-registry/pkg/sqlite"
)

var (
	packageName    = "prometheus"
	channels       = "preview"
	defaultChannel = "preview"

	bundlePath1 = "manifests/prometheus/0.14.0"

	bundleTag1 = rand.String(6)
	indexTag1  = rand.String(6)

	bundleImage = "quay.io/olmtest/e2e-bundle"
	indexImage1 = "quay.io/olmtest/e2e-index:" + indexTag1
)

type bundleLocation struct {
	image, path string
}

type bundleLocations []bundleLocation

func (bl bundleLocations) images() []string {
	images := make([]string, len(bl))
	for i, b := range bl {
		images[i] = b.image
	}

	return images
}

func inTemporaryBuildContext(f func() error) (rerr error) {
	td, err := ioutil.TempDir(".", "opm-")
	if err != nil {
		return err
	}
	err = copy.Copy("../../manifests", filepath.Join(td, "manifests"))
	if err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(td)
	if err != nil {
		return err
	}
	defer func() {
		err := os.Chdir(wd)
		if rerr == nil {
			rerr = err
		}
	}()
	return f()
}

func buildIndexWith(containerTool, fromIndexImage, toIndexImage string, bundleImages []string, mode registry.Mode, overwriteLatest bool) error {
	logger := logrus.WithFields(logrus.Fields{"bundles": bundleImages})
	indexAdder := indexer.NewIndexAdder(containertools.NewContainerTool(containerTool, containertools.NoneTool), containertools.NewContainerTool(containerTool, containertools.NoneTool), logger)

	request := indexer.AddToIndexRequest{
		Generate:          false,
		FromIndex:         fromIndexImage,
		BinarySourceImage: "",
		OutDockerfile:     "",
		Tag:               toIndexImage,
		Mode:              mode,
		Bundles:           bundleImages,
		Permissive:        false,
		Overwrite:         overwriteLatest,
	}

	return indexAdder.AddToIndex(request)
}

func pushWith(containerTool, image string) error {
	dockerpush := exec.Command(containerTool, "push", image)
	return dockerpush.Run()
}

func initialize() error {
	tmpDB, err := ioutil.TempFile("./", "index_tmp.db")
	if err != nil {
		return err
	}
	defer os.Remove(tmpDB.Name())

	db, err := sql.Open("sqlite3", tmpDB.Name())
	if err != nil {
		return err
	}
	defer db.Close()

	dbLoader, err := sqlite.NewSQLLiteLoader(db)
	if err != nil {
		return err
	}
	if err := dbLoader.Migrate(context.TODO()); err != nil {
		return err
	}

	loader := sqlite.NewSQLLoaderForDirectory(dbLoader, "downloaded")
	return loader.Populate()
}

var _ = Describe("opm", func() {
	IncludeSharedSpecs := func(containerTool string) {
		BeforeEach(func() {
			if dockerUsername == "" || dockerPassword == "" {
				Skip("registry credentials are not available")
			}

			dockerlogin := exec.Command(containerTool, "login", "-u", dockerUsername, "-p", dockerPassword, "quay.io")
			err := dockerlogin.Run()
			Expect(err).NotTo(HaveOccurred(), "Error logging into quay.io")
		})

		It("builds bundle and index images", func() {
			By("building bundles")
			bundles := bundleLocations{
				{bundleTag1, bundlePath1},
			}
			var err error
			for _, b := range bundles {
				err = inTemporaryBuildContext(func() error {
					return bundle.BuildFunc(b.path, "", b.image, containerTool, packageName, channels, defaultChannel, false)
				})
				Expect(err).NotTo(HaveOccurred())
			}

			By("pushing bundles")
			for _, b := range bundles {
				Expect(pushWith(containerTool, b.image)).NotTo(HaveOccurred())
			}

			By("building an index")
			err = buildIndexWith(containerTool, "", indexImage1, bundles[:2].images(), registry.ReplacesMode, false)
			Expect(err).NotTo(HaveOccurred())

			By("pushing an index")
			err = pushWith(containerTool, indexImage1)
			Expect(err).NotTo(HaveOccurred())
		})
	}

	Context("using docker", func() {
		IncludeSharedSpecs("docker")
	})

	Context("using podman", func() {
		IncludeSharedSpecs("podman")
	})
})
