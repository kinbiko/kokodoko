package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// Reads the given args, & validate that only one argument was given.
// Validate that the arg is a file, directory, or specifies a line number or line numbers in a file
// Validate that the deepest directory of this path is a git repository
// Validate that this git repository has a GitHub remote
// Find the current HEAD of this repository
// Generate link
// Put link in clipboard
func run(ctx context.Context, args []string) error {

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

	g := git{repoPath: repoPath}

	remoteURL, err := g.remoteURL()
	if err != nil {
		return fmt.Errorf("unable to get remote URL: %w", err)
	}
	hash, err := g.hash()
	if err != nil {
		return fmt.Errorf("unable to get vcs hash: %w", err)
	}
	repoRoot, err := g.repoRoot()
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
	// for reference:
	// https://github.com/kinbiko/dotfiles/blob/15fc22c0c5672e0f15f2ef7ea333bd620aa9965c/vimrc#L35
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

type git struct {
	// repoPath doesn't necessarily point to the root of the repository -- just
	// **some** directory in the repo.
	repoPath string
}

func (g *git) remoteURL() (string, error) {
	stdout, err := exec.Command("git", "-C", g.repoPath, "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", fmt.Errorf("unable to execute git command: %w", err)
	}

	url := string(stdout)
	url = strings.ReplaceAll(url, "git@github.com:", "https://github.com/")
	return strings.ReplaceAll(url, ".git\n", ""), nil
}

func (g *git) hash() (string, error) {
	stdout, err := exec.Command("git", "-C", g.repoPath, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("unable to execute git command: %w", err)
	}
	return strings.ReplaceAll(string(stdout), "\n", ""), nil
}

func (g *git) repoRoot() (string, error) {
	stdout, err := exec.Command("git", "-C", g.repoPath, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("unable to execute git command: %w", err)
	}
	return strings.ReplaceAll(string(stdout), "\n", ""), nil
}

type systemCalls interface {
	remoteURL() (string, error)
	hash() (string, error)
}
