package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kinbiko/bugsnag"

	"github.com/atotto/clipboard"
	"github.com/kinbiko/kokodoko"
)

// Globals are used to enable Bugsnag.
// Must be updated per release.
const appVersion = "0.2.0"

//nolint:gochecknoglobals // The following are intended to be set via ldflags, for developers.
var (
	APIKey       = "INJECT_ME"
	ReleaseStage = "development"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type git struct {
	kokodoko.O11y
}

// RemoteURL returns the remote URL, in its https:// format, of the inner-most
// repository that contains the directory of the given path.
func (g *git) RemoteURL(ctx context.Context, repoPath string) (string, error) {
	cmd := fmt.Sprintf("git -C %s config --get remote.origin.url", repoPath)
	ctx = g.WithMetadatum(ctx, "system calls", "Remote URL", cmd)
	output, err := g.call(ctx, cmd)
	if err != nil {
		return "", err
	}
	output = strings.ReplaceAll(output, "git@github.com:", "https://github.com/")
	return strings.ReplaceAll(output, ".git\n", ""), nil
}

// Hash returns the current commit HEAD of the inner-most repository that
// contains the directory of the given path.
func (g *git) Hash(ctx context.Context, repoPath string) (string, error) {
	cmd := fmt.Sprintf("git -C %s rev-parse HEAD", repoPath)
	ctx = g.WithMetadatum(ctx, "system calls", "Hash", cmd)
	output, err := g.call(ctx, cmd)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(output, "\n", ""), nil
}

// RepoRoot returns the root of the repository the inner-most repository that
// contains the directory of the given path.
func (g *git) RepoRoot(ctx context.Context, repoPath string) (string, error) {
	cmd := fmt.Sprintf("git -C %s rev-parse --show-toplevel", repoPath)
	ctx = g.WithMetadatum(ctx, "system calls", "RepoRoot", cmd)
	output, err := g.call(ctx, cmd)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(output, "\n", ""), nil
}

// AbsolutePath isn't really executing a Git thing, but returns the absolute
// path to the given relative path.
func (g *git) AbsolutePath(relative string) (string, error) {
	return filepath.Abs(relative)
}

func (g *git) call(ctx context.Context, cmd string) (string, error) {
	s := strings.Split(cmd, " ")
	stdout, err := exec.Command(s[0], s[1:]...).Output() //nolint:gosec // Hey, it's your system. Hack yourself if you want
	if err != nil {
		return "", g.Wrap(ctx, err, "unable to execute git command '%s'", cmd)
	}
	return string(stdout), nil
}

func run(ctx context.Context, args []string) error {
	o11y, teardown := makeO11y()
	defer teardown()

	snag := func(err error, msg string) error {
		err = o11y.Wrap(ctx, err, msg)
		if n, ok := o11y.(*bugsnag.Notifier); ok {
			n.Notify(ctx, err)
		}
		return err //nolint:wrapcheck // It *is* wrapped. Linter is just too dumb to know.
	}

	url, err := kokodoko.New(&git{O11y: o11y}, o11y, kokodoko.Config{}).Run(ctx, args)
	if err != nil {
		return snag(err, "error when generating URL")
	}

	err = clipboard.WriteAll(url)
	if err != nil {
		return snag(err, fmt.Sprintf("unable to copy url '%s' to clipboard", url))
	}

	fmt.Printf("Copied '%s' to the clipboard!\n", url)
	return nil
}

type noopO11y struct{}

func (n *noopO11y) Wrap(ctx context.Context, err error, msgAndFmtArgs ...interface{}) error {
	// A happy accident of good design :)
	return bugsnag.Wrap(ctx, err, msgAndFmtArgs...)
}
func (n *noopO11y) WithMetadatum(ctx context.Context, tab, key string, val interface{}) context.Context {
	return ctx
}

func makeO11y() (kokodoko.O11y, func()) {
	n, err := bugsnag.New(bugsnag.Configuration{
		APIKey:       APIKey,
		AppVersion:   appVersion,
		ReleaseStage: ReleaseStage,
	})
	// Intentionally ignoring the error here -- the use of the Bugsnag
	// integration is entirely opt-in.
	if err != nil {
		return &noopO11y{}, func() {}
	}
	return n, n.Close
}
