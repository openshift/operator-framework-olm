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
	"strings"
	"text/tabwriter"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/cmd/generic-autobumper/bumper"
	"k8s.io/test-infra/prow/config/secret"
	"k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/labels"
)

type mode string

const (
	summarize   mode = "summarize"
	synchronize mode = "synchronize"
	publish     mode = "publish"
)

type fetchMode string

const (
	https fetchMode = "https"
	ssh   fetchMode = "ssh"
)

const (
	githubOrg   = "openshift"
	githubRepo  = "operator-framework-olm"
	githubLogin = "openshift-bot"

	defaultPRAssignee = "openshift/openshift-team-operator-runtime,openshift/openshift-team-operator-ecosystem"

	defaultBaseBranch = "master"
)

type options struct {
	stagingDir       string
	commitFileOutput string
	commitFileInput  string
	mode             string
	logLevel         string
	centralRef       string
	fetchMode        string

	dryRun       bool
	githubLogin  string
	githubOrg    string
	githubRepo   string
	gitName      string
	gitEmail     string
	gitSignoff   bool
	assign       string
	selfApprove  bool
	prBaseBranch string

	flagutil.GitHubOptions
}

func (o *options) Bind(fs *flag.FlagSet) {
	fs.StringVar(&o.stagingDir, "staging-dir", "staging/", "Directory for staging repositories.")
	fs.StringVar(&o.mode, "mode", string(summarize), fmt.Sprintf("Operation mode. One of %s", []mode{summarize, synchronize, publish}))
	fs.StringVar(&o.commitFileOutput, "commits-output", "", "File to write commits data to after resolving what needs to be synced.")
	fs.StringVar(&o.commitFileInput, "commits-input", "", "File to read commits data from in order to drive sync process.")
	fs.StringVar(&o.logLevel, "log-level", logrus.InfoLevel.String(), "Logging level.")
	fs.StringVar(&o.centralRef, "central-ref", "origin/master", "Git ref for the central branch that will be updated, used as the base for determining what commits need to be cherry-picked.")
	fs.StringVar(&o.fetchMode, "fetch-mode", string(ssh), "Method to use for fetching from git remotes.")

	fs.BoolVar(&o.dryRun, "dry-run", true, "Whether to actually create the pull request with github client")
	fs.StringVar(&o.githubLogin, "github-login", githubLogin, "The GitHub username to use.")
	fs.StringVar(&o.githubOrg, "org", githubOrg, "The downstream GitHub org name.")
	fs.StringVar(&o.githubRepo, "repo", githubRepo, "The downstream GitHub repository name.")
	fs.StringVar(&o.gitName, "git-name", "", "The name to use on the git commit. Requires --git-email. If not specified, uses the system default.")
	fs.StringVar(&o.gitEmail, "git-email", "", "The email to use on the git commit. Requires --git-name. If not specified, uses the system default.")
	fs.BoolVar(&o.gitSignoff, "git-signoff", false, "Whether to signoff the commit. (https://git-scm.com/docs/git-commit#Documentation/git-commit.txt---signoff)")
	fs.StringVar(&o.assign, "assign", defaultPRAssignee, "The comma-delimited set of github usernames or group names to assign the created pull request to.")
	fs.BoolVar(&o.selfApprove, "self-approve", false, "Self-approve the PR by adding the `approved` and `lgtm` labels. Requires write permissions on the repo.")
	fs.StringVar(&o.prBaseBranch, "pr-base-branch", defaultBaseBranch, "The base branch to use for the pull request.")
	o.GitHubOptions.AddFlags(fs)
	o.GitHubOptions.AllowAnonymous = true
}

func (o *options) Validate() error {
	switch mode(o.mode) {
	case summarize, synchronize, publish:
	default:
		return fmt.Errorf("--mode must be one of %v", []mode{summarize, synchronize})
	}

	switch fetchMode(o.fetchMode) {
	case ssh, https:
	default:
		return fmt.Errorf("--fetch-mode must be one of %v", []fetchMode{https, ssh})
	}

	if _, err := logrus.ParseLevel(o.logLevel); err != nil {
		return fmt.Errorf("--log-level invalid: %w", err)
	}

	if mode(o.mode) == publish {
		if o.githubLogin == "" {
			return fmt.Errorf("--github-login is mandatory")
		}
		if (o.gitEmail == "") != (o.gitName == "") {
			return fmt.Errorf("--git-name and --git-email must be specified together")
		}
		if o.assign == "" {
			return fmt.Errorf("--assign is mandatory")
		}

		if err := o.GitHubOptions.Validate(o.dryRun); err != nil {
			return err
		}
	}
	return nil
}

func (o *options) GitCommitArgs() []string {
	var commitArgs []string
	if o.gitSignoff {
		commitArgs = append(commitArgs, "--signoff")
	}
	return commitArgs
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
		commits, err = detectNewCommits(ctx, logger.WithField("phase", "detect"), opts.stagingDir, opts.centralRef, fetchMode(opts.fetchMode))
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

	var missingCommits []commit
	for _, commit := range commits {
		commitLogger := logger.WithField("commit", commit.Hash)
		missing, err := isCommitMissing(ctx, commitLogger, opts.stagingDir, commit)
		if err != nil {
			commitLogger.WithError(err).Fatal("failed to determine if commit is missing")
		}
		if missing {
			missingCommits = append(missingCommits, commit)
		}
	}

	cherryPickAll := func() {
		if err := setCommitter(ctx, logger.WithField("phase", "setup"), opts.gitName, opts.gitEmail); err != nil {
			logger.WithError(err).Fatal("failed to set committer")
		}
		for i, commit := range missingCommits {
			commitLogger := logger.WithField("commit", commit.Hash)
			commitLogger.Infof("cherry-picking commit %d/%d", i+1, len(missingCommits))
			if err := cherryPick(ctx, commitLogger, commit, opts.GitCommitArgs()); err != nil {
				logger.WithError(err).Fatal("failed to cherry-pick commit")
			}
		}
	}

	if len(missingCommits) == 0 {
		logger.Info("Current repository state is up-to-date with upstream.")
		return
	}

	switch mode(opts.mode) {
	case summarize:
		writer := tabwriter.NewWriter(bumper.HideSecretsWriter{Delegate: os.Stdout, Censor: secret.Censor}, 0, 4, 2, ' ', 0)
		for _, commit := range missingCommits {
			if _, err := fmt.Fprintln(writer, commit.Date.Format(time.DateTime)+"\t"+"operator-framework/"+commit.Repo+"\t", commit.Hash+"\t"+commit.Author+"\t"+commit.Message); err != nil {
				logger.WithError(err).Error("failed to write output")
			}
		}
		if err := writer.Flush(); err != nil {
			logger.WithError(err).Error("failed to flush output")
		}
	case synchronize:
		cherryPickAll()
	case publish:
		cherryPickAll()
		gc, err := opts.GitHubOptions.GitHubClient(opts.dryRun)
		if err != nil {
			logrus.WithError(err).Fatal("error getting GitHub client")
		}
		gc.SetMax404Retries(0)

		stdout := bumper.HideSecretsWriter{Delegate: os.Stdout, Censor: secret.Censor}
		stderr := bumper.HideSecretsWriter{Delegate: os.Stderr, Censor: secret.Censor}

		remoteBranch := "synchronize-upstream"
		title := "Synchronize From Upstream Repositories"
		if err := bumper.MinimalGitPush(fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", opts.githubLogin,
			string(secret.GetTokenGenerator(opts.GitHubOptions.TokenPath)()), opts.githubLogin, opts.githubRepo),
			remoteBranch, stdout, stderr, opts.dryRun); err != nil {
			logrus.WithError(err).Fatal("Failed to push changes.")
		}

		var labelsToAdd []string
		if opts.selfApprove {
			logrus.Infof("Self-aproving PR by adding the %q and %q labels", labels.Approved, labels.LGTM)
			labelsToAdd = append(labelsToAdd, labels.Approved, labels.LGTM)
		}
		if err := bumper.UpdatePullRequestWithLabels(gc, opts.githubOrg, opts.githubRepo, title,
			getBody(commits, strings.Split(opts.assign, ",")), opts.githubLogin+":"+remoteBranch, opts.prBaseBranch, remoteBranch, true, labelsToAdd, opts.dryRun); err != nil {
			logrus.WithError(err).Fatal("PR creation failed.")
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

func detectNewCommits(ctx context.Context, logger *logrus.Entry, stagingDir, centralRef string, mode fetchMode) ([]commit, error) {
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
			centralRef,
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

	commits := map[string][]commit{}
	for repo, lastCommit := range lastCommits {
		var remote string
		switch mode {
		case ssh:
			remote = "git@github.com:operator-framework/" + repo
		case https:
			remote = "https://github.com/operator-framework/" + repo + ".git"
		}
		if _, err := runCommand(logger, exec.CommandContext(ctx,
			"git", "fetch",
			remote,
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
				infoCmd.Stdout = bumper.HideSecretsWriter{Delegate: &stdout, Censor: secret.Censor}
				infoCmd.Stderr = bumper.HideSecretsWriter{Delegate: &stderr, Censor: secret.Censor}
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
				if _, ok := commits[repo]; !ok {
					commits[repo] = []commit{}
				}
				commits[repo] = append(commits[repo], commit{
					Hash:    parts[0],
					Date:    committedTime,
					Author:  parts[2],
					Message: parts[3],
					Repo:    repo,
				})
			}
		}
	}
	// we would like to intertwine the commits from each upstream repository by date, while
	// keeping the order of commits from any one repository in the order they were committed in
	var orderedCommits []commit
	indices := map[string]int{}
	for repo := range commits {
		indices[repo] = 0
	}
	for {
		// find which repo's commit stack we should pop off to get the next earliest commit
		nextTime := time.Now()
		var nextRepo string
		for repo, index := range indices {
			if commits[repo][index].Date.Before(nextTime) {
				nextTime = commits[repo][index].Date
				nextRepo = repo
			}
		}

		// pop the commit, add it to our list and do housekeeping for our index records
		orderedCommits = append(orderedCommits, commits[nextRepo][indices[nextRepo]])
		if indices[nextRepo] == len(commits[nextRepo])-1 {
			delete(indices, nextRepo)
		} else {
			indices[nextRepo] += 1
		}

		if len(indices) == 0 {
			break
		}
	}

	// our ordered list is descending, but we need to cherry-pick from the oldest first
	var reversedCommits []commit
	for i := range orderedCommits {
		reversedCommits = append(reversedCommits, orderedCommits[len(orderedCommits)-i-1])
	}
	return reversedCommits, nil
}

func isCommitMissing(ctx context.Context, logger *logrus.Entry, stagingDir string, c commit) (bool, error) {
	output, err := runCommand(logger, exec.CommandContext(ctx,
		"git", "log",
		"-n", "1",
		"--grep", "Upstream-repository: "+c.Repo,
		"--grep", "Upstream-commit: "+c.Hash,
		"--all-match",
		"--pretty=%B",
		"--",
		filepath.Join(stagingDir, c.Repo),
	))
	if err != nil {
		return false, err
	}
	return len(output) == 0, nil
}

func setCommitter(ctx context.Context, logger *logrus.Entry, name string, email string) error {
	for field, value := range map[string]string{
		"user.name":  name,
		"user.email": email,
	} {
		output, err := runCommand(logger, exec.CommandContext(ctx,
			"git", "config",
			"--get", field,
		))
		if err != nil {
			return err
		}
		if len(output) == 0 {
			_, err := runCommand(logger, exec.CommandContext(ctx,
				"git", "config",
				"--add", field, value,
			))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func cherryPick(ctx context.Context, logger *logrus.Entry, c commit, commitArgs []string) error {
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
			"make", "generate-manifests",
		), os.Environ()...),
		exec.CommandContext(ctx,
			"git", "add",
			"staging/"+c.Repo,
			"vendor", "go.mod", "go.sum",
			"manifests", "pkg/manifests",
		),
		exec.CommandContext(ctx,
			"git", append([]string{"commit",
				"--amend", "--allow-empty", "--no-edit",
				"--trailer", "Upstream-repository: " + c.Repo,
				"--trailer", "Upstream-commit: " + c.Hash,
				"staging/" + c.Repo,
				"vendor", "go.mod", "go.sum",
				"manifests", "pkg/manifests"},
				commitArgs...)...,
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
	cmd.Stdout = bumper.HideSecretsWriter{Delegate: &output, Censor: secret.Censor}
	cmd.Stderr = bumper.HideSecretsWriter{Delegate: &output, Censor: secret.Censor}
	logger = logger.WithField("command", cmd.String())
	logger.Debug("running command")
	if err := cmd.Run(); err != nil {
		return output.String(), fmt.Errorf("failed to run command: %s: %w", output.String(), err)
	}
	logger.WithField("output", output.String()).Debug("ran command")
	return output.String(), nil
}

func withEnv(command *exec.Cmd, env ...string) *exec.Cmd {
	command.Env = append(command.Env, env...)
	return command
}

func getBody(commits []commit, assign []string) string {
	lines := []string{
		"The staging/ and vendor/ directories have been synchronized from the upstream repositories, pulling in the following commits:",
		"",
		"| Date | Commit | Author | Message |",
		"| -    | -      | -      | -       |",
	}
	for _, commit := range commits {
		lines = append(
			lines,
			fmt.Sprintf("|%s|[operator-framework/%s@%s](https://github.com/operator-framework/%s/commit/%s)|%s|%s|",
				commit.Date.Format(time.DateTime),
				commit.Repo,
				commit.Hash[0:7],
				commit.Repo,
				commit.Hash,
				commit.Author,
				commit.Message,
			),
		)
	}
	lines = append(lines, "", "This pull request is expected to merge without any human intervention. If tests are failing here, changes must land upstream to fix any issues so that future downstreaming efforts succeed.", "")
	for _, who := range assign {
		lines = append(lines, fmt.Sprintf("/cc @%s", who))
	}

	body := strings.Join(lines, "\n")

	if len(body) >= 65536 {
		body = body[:65530] + "..."
	}

	return body
}
