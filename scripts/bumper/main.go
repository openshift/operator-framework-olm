package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/sirupsen/logrus"
)

type mode string

const (
	summarize   mode = "summarize"
	synchronize mode = "synchronize"
)

type options struct {
	stagingDir       string
	commitFileOutput string
	commitFileInput  string
	mode             string
	logLevel         string
}

func (o *options) Bind(fs *flag.FlagSet) {
	fs.StringVar(&o.stagingDir, "staging-dir", "staging/", "Directory for staging repositories.")
	fs.StringVar(&o.mode, "mode", string(summarize), "Operation mode.")
	fs.StringVar(&o.commitFileOutput, "commits-output", "", "File to write commits data to after resolving what needs to be synced.")
	fs.StringVar(&o.commitFileInput, "commits-input", "", "File to read commits data from in order to drive sync process.")
	fs.StringVar(&o.logLevel, "log-level", logrus.InfoLevel.String(), "Logging level.")
}

func (o *options) Validate() error {
	switch mode(o.mode) {
	case summarize, synchronize:
	default:
		return fmt.Errorf("--mode must be one of %v", []mode{summarize, synchronize})
	}

	if _, err := logrus.ParseLevel(o.logLevel); err != nil {
		return fmt.Errorf("--log-level invalid: %w", err)
	}
	return nil
}

func main() {
	logger := logrus.New()
	opts := options{}
	opts.Bind(flag.CommandLine)
	flag.Parse()

	if err := opts.Validate(); err != nil {
		logger.WithError(err).Fatal("invalid options")
	}

	logLevel, _ := logrus.ParseLevel(opts.logLevel)
	logger.SetLevel(logLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var commits []commit
	var err error
	if opts.commitFileInput != "" {
		rawCommits, err := os.ReadFile(opts.commitFileInput)
		if err != nil {
			logrus.WithError(err).Fatal("could not read input file")
		}
		if err := json.Unmarshal(rawCommits, &commits); err != nil {
			logrus.WithError(err).Fatal("could not unmarshal input commits")
		}
	} else {
		commits, err = detectNewCommits(ctx, logger.WithField("phase", "detect"), opts.stagingDir)
		if err != nil {
			logger.WithError(err).Fatal("failed to detect commits")
		}
	}

	if opts.commitFileOutput != "" {
		commitsJson, err := json.Marshal(commits)
		if err != nil {
			logrus.WithError(err).Fatal("could not marshal commits")
		}
		if err := os.WriteFile(opts.commitFileOutput, commitsJson, 0666); err != nil {
			logrus.WithError(err).Fatal("could not write commits")
		}
	}

	switch mode(opts.mode) {
	case summarize:
		writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		for _, commit := range commits {
			if _, err := fmt.Fprintln(writer, commit.Date.Format(time.DateTime)+"\t"+"operator-framework/"+commit.Repo+"\t", commit.Hash+"\t"+commit.Author+"\t"+commit.Message); err != nil {
				logger.WithError(err).Error("failed to write output")
			}
		}
		if err := writer.Flush(); err != nil {
			logger.WithError(err).Error("failed to flush output")
		}
	case synchronize:
		for i, commit := range commits {
			commitLogger := logger.WithField("commit", commit.Hash)
			commitLogger.Infof("cherry-picking commit %d/%d", i+1, len(commits))
			if err := cherryPick(ctx, commitLogger, commit); err != nil {
				logger.WithError(err).Error("failed to cherry-pick commit")
				break
			}
		}
	}
}

type commit struct {
	Date    time.Time `json:"date"`
	Hash    string    `json:"hash,omitempty"`
	Author  string    `json:"author,omitempty"`
	Message string    `json:"message,omitempty"`
	Repo    string    `json:"repo,omitempty"`
}

var repoRegex = regexp.MustCompile(`Upstream-repository: ([^ ]+)\n`)
var commitRegex = regexp.MustCompile(`Upstream-commit: ([a-f0-9]+)\n`)

func detectNewCommits(ctx context.Context, logger *logrus.Entry, stagingDir string) ([]commit, error) {
	lastCommits := map[string]string{}
	if err := fs.WalkDir(os.DirFS(stagingDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil || !d.IsDir() {
			return nil
		}

		if path == "." {
			return nil
		}
		logger = logger.WithField("repo", path)
		logger.Debug("detecting commits")
		output, err := runCommand(logger, exec.CommandContext(ctx,
			"git", "log",
			"-n", "1",
			"--grep", "Upstream-repository: "+path,
			"--grep", "Upstream-commit",
			"--all-match",
			"--pretty=%B",
			"--",
			filepath.Join(stagingDir, path),
		))
		if err != nil {
			return err
		}
		var lastCommit string
		commitMatches := commitRegex.FindStringSubmatch(output)
		if len(commitMatches) > 0 {
			if len(commitMatches[0]) > 1 {
				lastCommit = string(commitMatches[1])
			}
		}
		if lastCommit != "" {
			logger.WithField("commit", lastCommit).Debug("found last commit synchronized with staging")
			lastCommits[path] = lastCommit
		}

		if path != "." {
			return fs.SkipDir
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk %s: %w", stagingDir, err)
	}

	var commits []commit
	for repo, lastCommit := range lastCommits {
		if _, err := runCommand(logger, exec.CommandContext(ctx,
			"git", "fetch",
			"git@github.com:operator-framework/"+repo,
			"master",
		)); err != nil {
			return nil, err
		}

		output, err := runCommand(logger, exec.CommandContext(ctx,
			"git", "log",
			"--pretty=%H",
			lastCommit+"...FETCH_HEAD",
		))
		if err != nil {
			return nil, err
		}

		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				infoCmd := exec.CommandContext(ctx,
					"git", "show",
					line,
					"--pretty=format:%H\u00A0%cI\u00A0%an\u00A0%s",
					"--quiet",
				)
				stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
				infoCmd.Stdout = &stdout
				infoCmd.Stderr = &stderr
				logger.WithField("command", infoCmd.String()).Debug("running command")
				if err := infoCmd.Run(); err != nil {
					return nil, fmt.Errorf("failed to run command: %s %s: %w", stdout.String(), stderr.String(), err)
				}
				parts := strings.Split(stdout.String(), "\u00A0")
				if len(parts) != 4 {
					return nil, fmt.Errorf("incorrect parts from git output: %v", stdout.String())
				}
				committedTime, err := time.Parse(time.RFC3339, parts[1])
				if err != nil {
					return nil, fmt.Errorf("invalid time %s: %w", parts[1], err)
				}
				commits = append(commits, commit{
					Hash:    parts[0],
					Date:    committedTime,
					Author:  parts[2],
					Message: parts[3],
					Repo:    repo,
				})
			}
		}
	}
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Date.Before(commits[j].Date)
	})
	return commits, nil
}

func cherryPick(ctx context.Context, logger *logrus.Entry, c commit) error {
	{
		output, err := runCommand(logger, exec.CommandContext(ctx,
			"git", "cherry-pick",
			"--allow-empty", "--keep-redundant-commits",
			"-Xsubtree=staging/"+c.Repo, c.Hash,
		))
		if err != nil {
			if strings.Contains(output, "vendor/modules.txt deleted in HEAD and modified in") {
				// we remove vendor directories for everything under staging/, but some of the upstream repos have them
				if _, err := runCommand(logger, exec.CommandContext(ctx,
					"git", "rm", "--cached", "-r", "--ignore-unmatch", "staging/"+c.Repo+"/vendor",
				)); err != nil {
					return err
				}
				if _, err := runCommand(logger, exec.CommandContext(ctx,
					"git", "cherry-pick", "--continue",
				)); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, cmd := range []*exec.Cmd{
		withEnv(exec.CommandContext(ctx,
			"go", "mod", "tidy",
		), os.Environ()...),
		withEnv(exec.CommandContext(ctx,
			"go", "mod", "vendor",
		), os.Environ()...),
		withEnv(exec.CommandContext(ctx,
			"go", "mod", "verify",
		), os.Environ()...),
		withEnv(exec.CommandContext(ctx,
			"make", "manifests", "OLM_VERSION=0.0.1-snapshot",
		), os.Environ()...),
		exec.CommandContext(ctx,
			"git", "add",
			"staging/"+c.Repo,
			"vendor", "go.mod", "go.sum",
			"manifests", "pkg/manifests",
		),
		exec.CommandContext(ctx,
			"git", "commit",
			"--amend", "--allow-empty", "--no-edit",
			"--trailer", "Upstream-repository: "+c.Repo,
			"--trailer", "Upstream-commit: "+c.Hash,
			"staging/"+c.Repo,
			"vendor", "go.mod", "go.sum",
			"manifests", "pkg/manifests",
		),
	} {
		if _, err := runCommand(logger, cmd); err != nil {
			return err
		}
	}

	return nil
}

func runCommand(logger *logrus.Entry, cmd *exec.Cmd) (string, error) {
	output := bytes.Buffer{}
	cmd.Stdout = &output
	cmd.Stderr = &output
	logger.WithField("command", cmd.String()).Debug("running command")
	if err := cmd.Run(); err != nil {
		return output.String(), fmt.Errorf("failed to run command: %s: %w", output.String(), err)
	}
	return output.String(), nil
}

func withEnv(command *exec.Cmd, env ...string) *exec.Cmd {
	command.Env = append(command.Env, env...)
	return command
}
