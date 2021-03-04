/*
Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/pkg/plugin"
)

type pluginError struct {
	error
	code int
}

// loadPlugins loads plugins into the command list.
//
// This follows a different pattern than the other commands because it has
// to inspect its environment and then add commands to the base command
// as it finds them.
func loadPlugins(baseCmd *cobra.Command, out io.Writer) {

	// If HELM_NO_PLUGINS is set to 1, do not load plugins.
	if os.Getenv("HELM_NO_PLUGINS") == "1" {
		return
	}

	found, err := findPlugins(settings.PluginsDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load plugins: %s", err)
		return
	}

	processParent := func(cmd *cobra.Command, args []string) ([]string, error) {
		k, u := manuallyProcessArgs(args)
		if err := cmd.Parent().ParseFlags(k); err != nil {
			return nil, err
		}
		return u, nil
	}

	// Now we create commands for all of these.
	for _, plug := range found {
		plug := plug
		md := plug.Metadata
		if md.Usage == "" {
			md.Usage = fmt.Sprintf("the %q plugin", md.Name)
		}

		c := &cobra.Command{
			Use:   md.Name,
			Short: md.Usage,
			Long:  md.Description,
			RunE: func(cmd *cobra.Command, args []string) error {
				u, err := processParent(cmd, args)
				if err != nil {
					return err
				}

				// Call setupEnv before PrepareCommand because
				// PrepareCommand uses os.ExpandEnv and expects the
				// setupEnv vars.
				plugin.SetupPluginEnv(settings, md.Name, plug.Dir)
				main, argv, prepCmdErr := plug.PrepareCommand(u)
				if prepCmdErr != nil {
					os.Stderr.WriteString(prepCmdErr.Error())
					return errors.Errorf("plugin %q exited with error", md.Name)
				}

				env := os.Environ()
				for k, v := range settings.EnvVars() {
					env = append(env, fmt.Sprintf("%s=%s", k, v))
				}

				prog := exec.Command(main, argv...)
				prog.Env = env
				prog.Stdin = os.Stdin
				prog.Stdout = out
				prog.Stderr = os.Stderr
				if err := prog.Run(); err != nil {
					if eerr, ok := err.(*exec.ExitError); ok {
						os.Stderr.Write(eerr.Stderr)
						status := eerr.Sys().(syscall.WaitStatus)
						return pluginError{
							error: errors.Errorf("plugin %q exited with error", md.Name),
							code:  status.ExitStatus(),
						}
					}
					return err
				}
				return nil
			},
			// This passes all the flags to the subcommand.
			DisableFlagParsing: true,
		}

		// TODO: Make sure a command with this name does not already exist.
		baseCmd.AddCommand(c)
	}
}

// manuallyProcessArgs processes an arg array, removing special args.
//
// Returns two sets of args: known and unknown (in that order)
func manuallyProcessArgs(args []string) ([]string, []string) {
	known := []string{}
	unknown := []string{}
	kvargs := []string{"--kube-context", "--namespace", "-n", "--kubeconfig", "--registry-config", "--repository-cache", "--repository-config"}
	knownArg := func(a string) bool {
		for _, pre := range kvargs {
			if strings.HasPrefix(a, pre+"=") {
				return true
			}
		}
		return false
	}

	isKnown := func(v string) string {
		for _, i := range kvargs {
			if i == v {
				return v
			}
		}
		return ""
	}

	for i := 0; i < len(args); i++ {
		switch a := args[i]; a {
		case "--debug":
			known = append(known, a)
		case isKnown(a):
			known = append(known, a)
			i++
			if i < len(args) {
				known = append(known, args[i])
			}
		default:
			if knownArg(a) {
				known = append(known, a)
				continue
			}
			unknown = append(unknown, a)
		}
	}
	return known, unknown
}

// findPlugins returns a list of YAML files that describe plugins.
func findPlugins(plugdirs string) ([]*plugin.Plugin, error) {
	found := []*plugin.Plugin{}
	// Let's get all UNIXy and allow path separators
	for _, p := range filepath.SplitList(plugdirs) {
		matches, err := plugin.LoadAll(p)
		if err != nil {
			return matches, err
		}
		found = append(found, matches...)
	}
	return found, nil
}
