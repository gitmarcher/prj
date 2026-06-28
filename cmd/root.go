package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "prj",
	Short: "Bootstrap multi-repository engineering workspaces",
	Long: `prj helps you create and manage multi-repository project workspaces.

It clones repositories, creates feature branches, collects project context,
and generates a CLAUDE.md file ready for AI-assisted development.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(statusCmd)
}
