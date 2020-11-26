package kokodoko

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/kinbiko/bugsnag"
)

// O11y exposes observability methods for monitoring this application.
type O11y interface {
	// TODO(https://github.com/kinbiko/bugsnag/issues/31): The bugsnag
	// dependency here is unfortunate:
	Wrap(ctx context.Context, err error, msgAndFmtArgs ...interface{}) *bugsnag.Error

	WithMetadatum(ctx context.Context, tab, key string, value interface{}) context.Context
}

// Kokodoko is the core application struct that executes everything.
type Kokodoko struct {
	sys  System
	o11y O11y
	cfg  Config
}

// System represents data-fetching methods that require system calls.
type System interface {
	// repoPath doesn't have to be the root of the repository -- any directory in the repo.
	RemoteURL(ctx context.Context, repoPath string) (string, error)
	// repoPath doesn't have to be the root of the repository -- any directory in the repo.
	Hash(ctx context.Context, repoPath string) (string, error)
	// repoPath doesn't have to be the root of the repository -- any directory in the repo.
	RepoRoot(ctx context.Context, repoPath string) (string, error)
}

// Config holds options that alters the behaviour of the app.
// Note: The author reserves the right to add fields to this struct without
// releasing a new major version. You should be fine as long as you name
// parameters in this struct.
type Config struct {
	// Intentionally empty.
	// Added in order to support
}

// New creates a new Kokodoko application based on the given dependencies and
// configuration.
func New(sys System, o11y O11y, cfg Config) *Kokodoko {
	return &Kokodoko{sys: sys, o11y: o11y, cfg: cfg}
}

// Run reads the given args, & validate that only one argument was given.
// Validate that the arg is a file, directory, or specifies a line number or line numbers in a file
// Validate that the deepest directory of this path is a git repository
// Validate that this git repository has a GitHub remote
// Find the current HEAD of this repository
// Generate link
// Put link in clipboard
func (k *Kokodoko) Run(ctx context.Context, args []string) error {
	pathCandidate, lines, err := readArg(args)
	if err != nil {
		return fmt.Errorf("argument error: %w", err)
	}

	info, err := os.Stat(pathCandidate)
	if os.IsNotExist(err) {
		return fmt.Errorf("no such file or directory '%s'", pathCandidate)
	}
	if info.IsDir() && lines != "" {
		return fmt.Errorf("'%s' is a directory, so line numbers don't make sense", pathCandidate)
	}
	// repoPath doesn't necessarily point to the root of the repository -- just
	// **some** directory in the repo.
	repoPath := pathCandidate
	if !info.IsDir() {
		repoPath = repoPath[:len(repoPath)-len(info.Name())]
	}

	remoteURL, err := k.sys.RemoteURL(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("unable to get remote URL: %w", err)
	}
	hash, err := k.sys.Hash(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("unable to get vcs hash: %w", err)
	}
	repoRoot, err := k.sys.RepoRoot(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("unable to get repository root directory: %w", err)
	}

	if lines != "" {
		// Turn "1-51" into "#L1-L52" like it is in GitHub URLs.
		lines = strings.ReplaceAll("#L"+lines, "-", "-L")
	}
	absolutePath, err := filepath.Abs(pathCandidate)
	if err != nil {
		return fmt.Errorf("unable to get absolute filepath to '%s': %w", pathCandidate, err)
	}
	filePathRelativeToGitRoot := absolutePath[strings.Index(absolutePath, repoRoot)+len(repoRoot):]
	// desired URL for reference:
	// https://github.com/kinbiko/dotfiles/blob/15fc22c0c5672e0f15f2ef7ea333bd620aa9965c/vimrc#L35-L52
	url := fmt.Sprintf("%s/blob/%s%s%s", remoteURL, hash, filePathRelativeToGitRoot, lines)

	if err = clipboard.WriteAll(url); err != nil {
		return fmt.Errorf("unable to copy url '%s' to clipboard: %w", url, err)
	}

	fmt.Printf("Copied '%s' to the clipboard!\n", url)
	return nil
}

func readArg(args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf(`no path given, did you mean "."?`)
	}
	if len(args) > 2 {
		return "", "", fmt.Errorf("only one or arguments expected")
	}

	if len(args) == 1 {
		return args[0], "", nil
	}
	return args[0], args[1], nil
}
