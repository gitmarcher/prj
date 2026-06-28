package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

// Meta holds all project metadata.
type Meta struct {
	Name      string    `yaml:"name"`
	Source    string    `yaml:"source"`
	Branch    string    `yaml:"branch"`
	Primary   []string  `yaml:"primary"`
	Secondary []string  `yaml:"secondary,omitempty"`
	Intent    string    `yaml:"intent"`
	Docs      []string  `yaml:"docs,omitempty"`
	Jira      []string  `yaml:"jira,omitempty"`
	Slack     string    `yaml:"slack,omitempty"`
	Notes     string    `yaml:"notes,omitempty"`
	CreatedAt time.Time `yaml:"created_at"`
}

// WorkspaceDir returns the full path of the project workspace directory.
func WorkspaceDir(workspace, name string) string {
	return filepath.Join(workspace, "prj-"+name)
}

// CreateWorkspace creates the workspace directory and returns its path.
func CreateWorkspace(workspace, name string) (string, error) {
	dir := WorkspaceDir(workspace, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("cannot create workspace %s: %w", dir, err)
	}
	return dir, nil
}

// WriteConfig writes .prj/config.yaml inside workspaceDir.
func WriteConfig(workspaceDir string, m *Meta) error {
	prjDir := filepath.Join(workspaceDir, ".prj")
	if err := os.MkdirAll(prjDir, 0o755); err != nil {
		return fmt.Errorf("cannot create .prj directory: %w", err)
	}

	data, err := marshalYAML(m)
	if err != nil {
		return fmt.Errorf("cannot marshal project config: %w", err)
	}

	dest := filepath.Join(prjDir, "config.yaml")
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("cannot write project config: %w", err)
	}
	return nil
}

func marshalYAML(v any) ([]byte, error) {
	var sb strings.Builder
	enc := yaml.NewEncoder(&sb)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}
