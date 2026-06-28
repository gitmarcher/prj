package gitstatus

import (
	"os/exec"
	"strconv"
	"strings"
)

// CurrentBranch returns the name of the currently checked-out branch.
func CurrentBranch(repoDir string) (string, error) {
	return output(repoDir, "rev-parse", "--abbrev-ref", "HEAD")
}

// DirtyCount returns the number of changed files in the working tree.
// Returns 0 for a clean tree.
func DirtyCount(repoDir string) (int, error) {
	out, err := output(repoDir, "status", "--porcelain")
	if err != nil {
		return 0, err
	}
	if out == "" {
		return 0, nil
	}
	return len(strings.Split(strings.TrimSpace(out), "\n")), nil
}

// AheadBehind returns commits ahead and behind the upstream tracking branch.
// If no upstream is configured, returns 0, 0, false (noUpstream=false means tracking exists;
// the third return value noUpstream is true when there is no tracking branch).
func AheadBehind(repoDir string) (ahead, behind int, noUpstream bool, err error) {
	out, err := output(repoDir, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		// exit 128 means no upstream configured
		return 0, 0, true, nil
	}
	parts := strings.Fields(out)
	if len(parts) != 2 {
		return 0, 0, false, nil
	}
	ahead, _ = strconv.Atoi(parts[0])
	behind, _ = strconv.Atoi(parts[1])
	return ahead, behind, false, nil
}

func output(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
