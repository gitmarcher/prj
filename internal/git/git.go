package git

import (
	"fmt"
	"os"
	"os/exec"
)

// Clone clones repoURL into destDir, streaming output to stdout/stderr.
func Clone(repoURL, destDir string) error {
	return run(destDir, "git", "clone", repoURL, destDir)
}

// CreateBranch creates and checks out branchName in repoDir.
// If the branch already exists, it checks it out instead.
func CreateBranch(repoDir, branchName string) error {
	err := runIn(repoDir, "git", "checkout", "-b", branchName)
	if err == nil {
		return nil
	}
	// Branch already exists — just check it out.
	return runIn(repoDir, "git", "checkout", branchName)
}

// DefaultBranch returns the default branch name (HEAD symref on the remote).
// Falls back to "main" on any error.
func DefaultBranch(repoDir string) string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "main"
	}
	// refs/remotes/origin/main -> main
	ref := string(out)
	for i := len(ref) - 1; i >= 0; i-- {
		if ref[i] == '/' {
			return ref[i+1 : len(ref)-1] // trim trailing newline
		}
	}
	return "main"
}

func run(destDir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", name, args, err)
	}
	return nil
}

func runIn(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", name, args, err)
	}
	return nil
}
