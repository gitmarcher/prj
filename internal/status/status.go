package status

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gitmarcher/prj/internal/gh"
	"github.com/gitmarcher/prj/internal/gitstatus"
	"github.com/gitmarcher/prj/internal/project"
)

type RepoStatus struct {
	Name        string
	Role        string // "primary" | "secondary"
	Branch      string
	DirtyCount  int
	Ahead       int
	Behind      int
	NoUpstream  bool
	PR          *gh.PR
	CI          *gh.Run
	Err         error
}

type KnowledgeFile struct {
	Name    string
	Exists  bool
	ModTime time.Time
	IsStale bool
}

type ProjectStatus struct {
	Meta      *project.Meta
	Repos     []*RepoStatus
	Knowledge []*KnowledgeFile
}

// Collect gathers status for all repos in the project concurrently.
func Collect(workspaceDir string, meta *project.Meta, knowledgeAgeDays int) *ProjectStatus {
	type entry struct {
		name string
		role string
	}

	entries := make([]entry, 0, len(meta.Primary)+len(meta.Secondary))
	for _, r := range meta.Primary {
		entries = append(entries, entry{r, "primary"})
	}
	for _, r := range meta.Secondary {
		entries = append(entries, entry{r, "secondary"})
	}

	ghOK := gh.IsInstalled()
	repos := make([]*RepoStatus, len(entries))
	var wg sync.WaitGroup
	for i, e := range entries {
		wg.Add(1)
		go func(idx int, name, role string) {
			defer wg.Done()
			repos[idx] = collectRepo(filepath.Join(workspaceDir, name), name, role, meta.Branch, ghOK)
		}(i, e.name, e.role)
	}
	wg.Wait()

	knowledge := collectKnowledge(workspaceDir, knowledgeAgeDays)

	return &ProjectStatus{
		Meta:      meta,
		Repos:     repos,
		Knowledge: knowledge,
	}
}

func collectRepo(repoDir, name, role, configuredBranch string, ghOK bool) *RepoStatus {
	rs := &RepoStatus{Name: name, Role: role}

	if _, err := os.Stat(repoDir); err != nil {
		rs.Err = fmt.Errorf("directory not found: %s", repoDir)
		return rs
	}

	branch, err := gitstatus.CurrentBranch(repoDir)
	if err != nil {
		rs.Err = fmt.Errorf("git error: %w", err)
		return rs
	}
	rs.Branch = branch

	dirty, err := gitstatus.DirtyCount(repoDir)
	if err != nil {
		rs.Err = fmt.Errorf("git error: %w", err)
		return rs
	}
	rs.DirtyCount = dirty

	ahead, behind, noUpstream, err := gitstatus.AheadBehind(repoDir)
	if err != nil {
		rs.Err = fmt.Errorf("git error: %w", err)
		return rs
	}
	rs.Ahead = ahead
	rs.Behind = behind
	rs.NoUpstream = noUpstream

	if ghOK {
		rs.PR, _ = gh.GetPR(repoDir, branch)
		if rs.PR != nil {
			rs.CI, _ = gh.GetLatestRun(repoDir, branch)
		}
	}

	return rs
}

func collectKnowledge(workspaceDir string, ageDays int) []*KnowledgeFile {
	names := []string{"project.md", "decisions.md", "open_questions.md"}
	files := make([]*KnowledgeFile, len(names))
	threshold := time.Duration(ageDays) * 24 * time.Hour
	prjDir := filepath.Join(workspaceDir, ".prj")

	for i, name := range names {
		kf := &KnowledgeFile{Name: name}
		info, err := os.Stat(filepath.Join(prjDir, name))
		if err == nil {
			kf.Exists = true
			kf.ModTime = info.ModTime()
			if ageDays > 0 {
				kf.IsStale = time.Since(kf.ModTime) > threshold
			}
		}
		files[i] = kf
	}
	return files
}
