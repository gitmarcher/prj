package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

// RepoAddition records a repository that was added after initial project creation.
type RepoAddition struct {
	Name    string    `yaml:"name"`
	Role    string    `yaml:"role"`
	Reason  string    `yaml:"reason"`
	AddedAt time.Time `yaml:"added_at"`
}

// Meta holds all project metadata.
type Meta struct {
	Name      string         `yaml:"name"`
	Source    string         `yaml:"source"`
	Branch    string         `yaml:"branch"`
	Primary   []string       `yaml:"primary"`
	Secondary []string       `yaml:"secondary,omitempty"`
	Intent    string         `yaml:"intent"`
	Docs      []string       `yaml:"docs,omitempty"`
	Jira      []string       `yaml:"jira,omitempty"`
	Slack     string         `yaml:"slack,omitempty"`
	Notes     string         `yaml:"notes,omitempty"`
	Additions []RepoAddition `yaml:"additions,omitempty"`
	CreatedAt time.Time      `yaml:"created_at"`
}

// HasRepo returns true if name already exists in Primary or Secondary.
func (m *Meta) HasRepo(name string) bool {
	for _, r := range m.Primary {
		if r == name {
			return true
		}
	}
	for _, r := range m.Secondary {
		if r == name {
			return true
		}
	}
	return false
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

// FindRoot walks up from start looking for a .prj/config.yaml directory marker.
func FindRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".prj", "config.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a prj workspace — no .prj/config.yaml found\nRun this command from within your project directory.")
		}
		dir = parent
	}
}

// LoadMeta reads and parses .prj/config.yaml from workspaceDir.
func LoadMeta(workspaceDir string) (*Meta, error) {
	data, err := os.ReadFile(filepath.Join(workspaceDir, ".prj", "config.yaml"))
	if err != nil {
		return nil, fmt.Errorf("cannot read .prj/config.yaml: %w", err)
	}
	var m Meta
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("cannot parse .prj/config.yaml: %w", err)
	}
	return &m, nil
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
