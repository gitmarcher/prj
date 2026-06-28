package gh

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type PR struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	State   string `json:"state"`   // OPEN, MERGED, CLOSED
	IsDraft bool   `json:"isDraft"`
}

type Run struct {
	Status       string `json:"status"`       // queued, in_progress, completed
	Conclusion   string `json:"conclusion"`   // success, failure, cancelled, ...
	WorkflowName string `json:"workflowName"`
}

// IsInstalled returns true if the gh CLI is available in PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// GetPR returns the most recent PR for branch in the repo at repoDir.
// Returns nil, nil when no PR exists.
func GetPR(repoDir, branch string) (*PR, error) {
	out, err := run(repoDir, "pr", "list",
		"--head", branch,
		"--state", "all",
		"--limit", "1",
		"--json", "number,title,url,state,isDraft",
	)
	if err != nil {
		return nil, err
	}
	var prs []PR
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return nil, fmt.Errorf("parsing PR list: %w", err)
	}
	if len(prs) == 0 {
		return nil, nil
	}
	return &prs[0], nil
}

// GetLatestRun returns the latest CI run for branch in the repo at repoDir.
// Returns nil, nil when no runs exist.
func GetLatestRun(repoDir, branch string) (*Run, error) {
	out, err := run(repoDir, "run", "list",
		"--branch", branch,
		"--limit", "1",
		"--json", "status,conclusion,workflowName",
	)
	if err != nil {
		return nil, err
	}
	var runs []Run
	if err := json.Unmarshal([]byte(out), &runs); err != nil {
		return nil, fmt.Errorf("parsing run list: %w", err)
	}
	if len(runs) == 0 {
		return nil, nil
	}
	return &runs[0], nil
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh %s: %s", strings.Join(args[:2], " "), strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
