// Package kokodoko exposes an API for fetching GitHub permalinks given a path to a local file.
package kokodoko

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// O11y exposes observability methods for monitoring this application.
type O11y interface {
	Wrap(ctx context.Context, err error, msgAndFmtArgs ...interface{}) error

	WithMetadatum(ctx context.Context, tab, key string, value interface{}) context.Context
}

// Kokodoko is the core application struct that executes everything.
type Kokodoko struct {
	cfg  Config
	sys  System
	o11y O11y
}

// System represents data-fetching methods that require system calls.
type System interface {
	// repoPath doesn't have to be the root of the repository -- any directory in the repo.
	RemoteURL(ctx context.Context, repoPath string) (string, error)
	// repoPath doesn't have to be the root of the repository -- any directory in the repo.
	Hash(ctx context.Context, repoPath string) (string, error)
	// repoPath doesn't have to be the root of the repository -- any directory in the repo.
	RepoRoot(ctx context.Context, repoPath string) (string, error)
	AbsolutePath(relative string) (string, error)
}

// Config holds options that alters the behavior of the app.
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

// Run reads the given args, performs some validation, makes some git syscalls,
// and returns the URL corresponding to the file (and line number(s))
// identified in the args.
// First arg is interpreted as a file path, whereas the second optional
// argument is a line number or a line number range in the form "12-51".
// nolint:funlen,gocyclo,cyclop // I rarely disagree with functions being too
// long, but in this case it *is* actually easier to read since it's all very
// left-margin aligned.
func (k *Kokodoko) Run(ctx context.Context, args []string) (string, error) {
	ctx = k.o11y.WithMetadatum(ctx, "app", "arguments", args)
	candidatePath, candidateLines, err := k.readArg(ctx, args)
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "candidate path", candidatePath)
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "candidate lines", candidateLines)
	if err != nil {
		return "", k.o11y.Wrap(ctx, err, "argument error")
	}

	info, err := os.Stat(candidatePath)
	if os.IsNotExist(err) {
		return "", k.o11y.Wrap(ctx, nil, "no such file or directory '%s'", candidatePath)
	}
	isDir := info.IsDir()
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "is directory", isDir)
	if isDir && candidateLines != "" {
		return "", k.o11y.Wrap(ctx, nil, "'%s' is a directory, so line numbers don't make sense", candidatePath)
	}
	// repoPath doesn't necessarily point to the root of the repository -- just
	// **some** directory in the repo.
	repoPath := candidatePath
	if !isDir {
		repoPath = repoPath[:len(repoPath)-len(info.Name())]
	}
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "repo path", repoPath)

	remoteURL, err := k.sys.RemoteURL(ctx, repoPath)
	if err != nil {
		return "", k.o11y.Wrap(ctx, err, "unable to get remote URL")
	}
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "remote URL", remoteURL)

	hash, err := k.sys.Hash(ctx, repoPath)
	if err != nil {
		return "", k.o11y.Wrap(ctx, err, "unable to get commit hash")
	}
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "commit hash", hash)

	repoRoot, err := k.sys.RepoRoot(ctx, repoPath)
	if err != nil {
		return "", k.o11y.Wrap(ctx, err, "unable to get repository root directory")
	}
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "repo root", repoRoot)

	lines := candidateLines
	if candidateLines != "" {
		// Turn "1-51" into "#L1-L52" like it is in GitHub URLs.
		lines = strings.ReplaceAll("#L"+lines, "-", "-L")
	}
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "lines", lines)

	absolutePath, err := k.sys.AbsolutePath(candidatePath)
	if err != nil {
		return "", k.o11y.Wrap(ctx, err, "unable to get absolute filepath to '%s'", candidatePath)
	}
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "absolute path", absolutePath)

	filePathRelativeToGitRoot := absolutePath[strings.Index(absolutePath, repoRoot)+len(repoRoot):]
	ctx = k.o11y.WithMetadatum(ctx, "candidate", "file path relative to git root", filePathRelativeToGitRoot)
	// desired URL for reference:
	// https://github.com/kinbiko/dotfiles/blob/15fc22c0c5672e0f15f2ef7ea333bd620aa9965c/vimrc#L35-L52
	url := fmt.Sprintf("%s/blob/%s%s%s", remoteURL, hash, filePathRelativeToGitRoot, lines)
	k.o11y.WithMetadatum(ctx, "candidate", "url", url)

	// Remove any whitespace that still exist in the string for whatever reason
	return strings.Join(strings.Fields(url), ""), nil
}

func (k *Kokodoko) readArg(ctx context.Context, args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", k.o11y.Wrap(ctx, nil, `no path given, did you mean "."?`)
	}
	if len(args) > 2 { //nolint:gomnd // Too keen
		return "", "", k.o11y.Wrap(ctx, nil, `at most two arguments expected`)
	}

	if len(args) == 1 {
		return args[0], "", nil
	}
	return args[0], args[1], nil
}
