package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gitmarcher/prj/internal/config"
	"github.com/gitmarcher/prj/internal/project"
	"github.com/gitmarcher/prj/internal/render"
	"github.com/gitmarcher/prj/internal/status"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show consolidated status for the current project workspace",
	RunE:  runStatus,
}

func runStatus(_ *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
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

	ps := status.Collect(root, meta, cfg.KnowledgeAgeDays)
	render.Dashboard(ps, true)
	return nil
}
