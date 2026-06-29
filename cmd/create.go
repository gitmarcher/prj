package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gitmarcher/prj/internal/claude"
	"github.com/gitmarcher/prj/internal/config"
	"github.com/gitmarcher/prj/internal/git"
	"github.com/gitmarcher/prj/internal/project"
	"github.com/gitmarcher/prj/internal/prompt"
)

var (
	flagSource    string
	flagPrimary   []string
	flagSecondary []string
	flagBranch    string
)

var createCmd = &cobra.Command{
	Use:   "create <project-name>",
	Short: "Create a new multi-repository project workspace",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringVar(&flagSource, "src", "", "Source key defined in config (e.g. carousell)")
	createCmd.Flags().StringSliceVarP(&flagPrimary, "p", "p", nil, "Primary repositories, comma-separated (feature branch will be created)")
	createCmd.Flags().StringSliceVarP(&flagSecondary, "s", "s", nil, "Secondary repositories, comma-separated (cloned, no branch change)")
	createCmd.Flags().StringVarP(&flagBranch, "b", "b", "", "Feature branch name to create on primary repos")

	_ = createCmd.MarkFlagRequired("src")
	_ = createCmd.MarkFlagRequired("b")
}

func runCreate(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	if len(flagPrimary) == 0 && len(flagSecondary) == 0 {
		return fmt.Errorf("at least one repository must be specified via --p or --s")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Validate source exists
	if _, ok := cfg.Sources[flagSource]; !ok {
		return fmt.Errorf("source %q not found; add it to ~/.config/prj/config.yaml", flagSource)
	}

	// 1. Create workspace directory
	workspaceDir, err := project.CreateWorkspace(cfg.Workspace, projectName)
	if err != nil {
		return err
	}
	fmt.Printf("\n==> Workspace: %s\n\n", workspaceDir)

	// 2. Clone all repositories
	allRepos := append(flagPrimary, flagSecondary...)
	for _, repo := range allRepos {
		cloneURL, err := cfg.ResolveCloneURL(flagSource, repo)
		if err != nil {
			return err
		}
		destDir := filepath.Join(workspaceDir, repo)
		if _, err := os.Stat(destDir); err == nil {
			fmt.Printf("  -- %s already exists, skipping clone\n", repo)
			continue
		}
		fmt.Printf("==> Cloning %s\n", cloneURL)
		if err := git.Clone(cloneURL, destDir); err != nil {
			return fmt.Errorf("cloning %s: %w", repo, err)
		}
	}

	// 3. Create feature branch on primary repositories
	if flagBranch != "" {
		for _, repo := range flagPrimary {
			repoDir := filepath.Join(workspaceDir, repo)
			fmt.Printf("\n==> Creating branch %q in %s\n", flagBranch, repo)
			if err := git.CreateBranch(repoDir, flagBranch); err != nil {
				return fmt.Errorf("creating branch in %s: %w", repo, err)
			}
		}
	}

	// 4. Interactive questions
	fmt.Println("\n==> Project context\n")

	intent, err := prompt.AskLongText("Project intent")
	if err != nil {
		return err
	}

	doneState, err := prompt.AskLongText("How will you know it's done?")
	if err != nil {
		return err
	}

	docs, err := prompt.AskOptionalMultiline("Source documentation (URLs or file paths)")
	if err != nil {
		return err
	}

	jiraLine, err := prompt.AskOptional("Jira ticket(s) (comma-separated)")
	if err != nil {
		return err
	}
	var jira []string
	for _, t := range strings.Split(jiraLine, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			jira = append(jira, t)
		}
	}

	slack, err := prompt.AskOptional("Slack channel")
	if err != nil {
		return err
	}

	notes, err := prompt.AskOptional("Additional notes")
	if err != nil {
		return err
	}

	// 5. Write project metadata
	meta := &project.Meta{
		Name:      projectName,
		Source:    flagSource,
		Branch:    flagBranch,
		Primary:   flagPrimary,
		Secondary: flagSecondary,
		Intent:    intent,
		DoneState: doneState,
		Docs:      docs,
		Jira:      jira,
		Slack:     slack,
		Notes:     notes,
		CreatedAt: time.Now().UTC(),
	}

	if err := project.WriteConfig(workspaceDir, meta); err != nil {
		return err
	}

	// 6. Generate CLAUDE.md and install skills
	if err := claude.WriteContextFile(workspaceDir, meta); err != nil {
		return err
	}
	if err := claude.WriteSkill(workspaceDir); err != nil {
		return err
	}

	fmt.Printf("\n==> Done! Project workspace ready at:\n    %s\n\n", workspaceDir)
	return nil
}
