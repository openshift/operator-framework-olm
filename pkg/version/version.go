package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// OLMVersion indicates what version of OLM the binary belongs to
	OLMVersion string

	// GitCommit indicates which git commit the binary was built from
	GitCommit string
	// buildDate in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
	buildDate string
)

type Version struct {
	OPMVersion string `json:"opmVersion"`
	GitCommit  string `json:"gitCommit"`
	BuildDate  string `json:"buildDate"`
	GoOs       string `json:"goOs"`
	GoArch     string `json:"goArch"`
}

// String returns a pretty string concatenation of OLMVersion and GitCommit
func String() string {
	return fmt.Sprintf("OLM version: %s\ngit commit: %s\n", OLMVersion, GitCommit)
}

// Full returns a hyphenated concatenation of just OLMVersion and GitCommit
func Full() string {
	return fmt.Sprintf("%s-%s", OLMVersion, GitCommit)
}

func getVersion() Version {
	return Version{
		OPMVersion: OLMVersion,
		GitCommit:  GitCommit,
		BuildDate:  buildDate,
		GoOs:       runtime.GOOS,
		GoArch:     runtime.GOARCH,
	}
}

func (v Version) Print() {
	fmt.Printf("Version: %#v\n", v)
}

func AddCommand(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print command and exit",
		Long:    `Print command version`,
		Example: `kubebuilder version`,
		Run:     runVersion,
	}

	parent.AddCommand(cmd)
}

func runVersion(_ *cobra.Command, _ []string) {
	getVersion().Print()
}
