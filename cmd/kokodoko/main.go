package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kinbiko/bugsnag"
	"github.com/kinbiko/kokodoko"
)

var (
	APIKey       = os.Getenv("KOKODOKO_BUGSNAG_API_KEY")
	AppVersion   = "INJECT ME"
	ReleaseStage = "development"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

/*
Wishlist before releasing:

- Wrap all errors in bugsnag.Wrap
- Tests
- Linting
- Bugsnag Integration
- Code of conduct
- Contributing guideline
- Issue templates
- README badges
	- Build status
	- License
	- Go report card
	- Docs
*/

type git struct {
	kokodoko.O11y
}

// RemoteURL returns the remote URL, in its https:// format, of the inner-most
// repository that contains the directory of the given path.
func (g *git) RemoteURL(ctx context.Context, repoPath string) (string, error) {
	cmd := fmt.Sprintf("git -C %s config --get remote.origin.url", repoPath)
	ctx = g.WithMetadatum(ctx, "system calls", "Remote URL", cmd)
	s := strings.Split(cmd, " ")
	stdout, err := exec.Command(s[0], s[1:]...).Output()
	if err != nil {
		return "", fmt.Errorf("unable to execute git command '%s': %w", cmd, err)
	}

	url := string(stdout)
	url = strings.ReplaceAll(url, "git@github.com:", "https://github.com/")
	return strings.ReplaceAll(url, ".git\n", ""), nil
}

// Hash returns the current commit HEAD of the inner-most repository that
// contains the directory of the given path.
func (g *git) Hash(ctx context.Context, repoPath string) (string, error) {
	cmd := fmt.Sprintf("git -C %s rev-parse HEAD", repoPath)
	ctx = g.WithMetadatum(ctx, "system calls", "Hash", cmd)
	s := strings.Split(cmd, " ")
	stdout, err := exec.Command(s[0], s[1:]...).Output()
	if err != nil {
		return "", fmt.Errorf("unable to execute git command '%s': %w", cmd, err)
	}
	return strings.ReplaceAll(string(stdout), "\n", ""), nil
}

// RepoRoot returns the root of the repository the inner-most repository that
// contains the directory of the given path.
func (g *git) RepoRoot(ctx context.Context, repoPath string) (string, error) {
	cmd := fmt.Sprintf("git -C %s rev-parse --show-toplevel", repoPath)
	ctx = g.WithMetadatum(ctx, "system calls", "Hash", cmd)
	s := strings.Split(cmd, " ")
	stdout, err := exec.Command(s[0], s[1:]...).Output()
	if err != nil {
		return "", fmt.Errorf("unable to execute git command '%s': %w", cmd, err)
	}
	return strings.ReplaceAll(string(stdout), "\n", ""), nil
}

func run(ctx context.Context, args []string) error {
	o11y, close := makeO11y()
	defer close()
	g := &git{O11y: o11y}
	app := kokodoko.New(g, o11y, kokodoko.Config{})
	return app.Run(ctx, args)
}

type noopO11y struct{}

func (n *noopO11y) Wrap(ctx context.Context, err error, msgAndFmtArgs ...interface{}) *bugsnag.Error {
	// A happy accident of good design :)
	return bugsnag.Wrap(ctx, err, msgAndFmtArgs...)
}
func (n *noopO11y) WithMetadatum(ctx context.Context, tab, key string, val interface{}) context.Context {
	return ctx
}

func makeO11y() (kokodoko.O11y, func()) {
	n, err := bugsnag.New(bugsnag.Configuration{
		APIKey:       APIKey,
		AppVersion:   AppVersion,
		ReleaseStage: ReleaseStage,
	})
	// Intentionally ignoring the error here -- the use of the Bugsnag
	// integration is entirely opt-in.
	if err != nil {
		return &noopO11y{}, func() {}
	}
	return n, n.Close
}
