package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/gitmarcher/prj/internal/claude"
	"github.com/gitmarcher/prj/internal/config"
	"github.com/gitmarcher/prj/internal/git"
	"github.com/gitmarcher/prj/internal/project"
	"github.com/gitmarcher/prj/internal/prompt"
)

var (
	repoAddPrimary   []string
	repoAddSecondary []string
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repositories within a project workspace",
}

var repoAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add one or more repositories to the current project",
	RunE:  runRepoAdd,
}

func init() {
	repoAddCmd.Flags().StringArrayVarP(&repoAddPrimary, "p", "p", nil, "Primary repositories to add (feature branch will be created)")
	repoAddCmd.Flags().StringArrayVarP(&repoAddSecondary, "s", "s", nil, "Secondary repositories to add (reference only)")
	repoCmd.AddCommand(repoAddCmd)
}

func runRepoAdd(_ *cobra.Command, _ []string) error {
	if len(repoAddPrimary) > 0 && len(repoAddSecondary) > 0 {
		return fmt.Errorf("only one repository type (--p or --s) may be specified per invocation")
	}
	if len(repoAddPrimary) == 0 && len(repoAddSecondary) == 0 {
		return fmt.Errorf("specify repositories with --p or --s")
	}

	repos, role := repoAddPrimary, "primary"
	if len(repoAddSecondary) > 0 {
		repos, role = repoAddSecondary, "secondary"
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := project.FindRoot(cwd)
	if err != nil {
		return err
	}
	meta, err := project.LoadMeta(root)
	if err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	added := 0
	for _, repo := range repos {
		if meta.HasRepo(repo) {
			fmt.Printf("  -- %s is already in this project, skipping\n", repo)
			continue
		}

		cloneURL, err := cfg.ResolveCloneURL(meta.Source, repo)
		if err != nil {
			return err
		}

		destDir := filepath.Join(root, repo)
		if _, err := os.Stat(destDir); err == nil {
			fmt.Printf("  -- %s directory already exists, skipping clone\n", repo)
		} else {
			fmt.Printf("\n==> Cloning %s\n", cloneURL)
			if err := git.Clone(cloneURL, destDir); err != nil {
				return fmt.Errorf("cloning %s: %w", repo, err)
			}
		}

		if role == "primary" {
			fmt.Printf("==> Creating branch %q in %s\n", meta.Branch, repo)
			if err := git.CreateBranch(destDir, meta.Branch); err != nil {
				return fmt.Errorf("creating branch in %s: %w", repo, err)
			}
		}

		reason, err := prompt.Ask("\nWhy is this repository being added?")
		if err != nil {
			return err
		}

		if role == "primary" {
			meta.Primary = append(meta.Primary, repo)
		} else {
			meta.Secondary = append(meta.Secondary, repo)
		}
		meta.Additions = append(meta.Additions, project.RepoAddition{
			Name:    repo,
			Role:    role,
			Reason:  reason,
			AddedAt: time.Now().UTC(),
		})

		added++
	}

	if added == 0 {
		fmt.Println("Nothing to add.")
		return nil
	}

	if err := project.WriteConfig(root, meta); err != nil {
		return err
	}
	if err := claude.WriteContextFile(root, meta); err != nil {
		return err
	}

	fmt.Printf("\n==> Updated .prj/config.yaml and CLAUDE.md\n")
	fmt.Printf("    Project knowledge is now out of date.\n")
	fmt.Printf("    Open Claude Code and run /project-analyzer to refresh.\n\n")
	return nil
}
