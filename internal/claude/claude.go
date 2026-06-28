package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gitmarcher/prj/internal/project"
)

// WriteContextFile generates CLAUDE.md at the project workspace root.
func WriteContextFile(workspaceDir string, m *project.Meta) error {
	content := buildContent(m)
	dest := filepath.Join(workspaceDir, "CLAUDE.md")
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return fmt.Errorf("cannot write CLAUDE.md: %w", err)
	}
	return nil
}

func buildContent(m *project.Meta) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", m.Name))

	sb.WriteString("## Intent\n\n")
	sb.WriteString(strings.TrimSpace(m.Intent))
	sb.WriteString("\n\n")

	sb.WriteString("## Repositories\n\n")
	sb.WriteString("### Primary\n\n")
	for _, r := range m.Primary {
		sb.WriteString(fmt.Sprintf("- `%s` (branch: `%s`)\n", r, m.Branch))
	}
	if len(m.Secondary) > 0 {
		sb.WriteString("\n### Secondary\n\n")
		for _, r := range m.Secondary {
			sb.WriteString(fmt.Sprintf("- `%s`\n", r))
		}
	}
	sb.WriteString("\n")

	if len(m.Docs) > 0 {
		sb.WriteString("## Documentation\n\n")
		for _, d := range m.Docs {
			sb.WriteString(fmt.Sprintf("- %s\n", d))
		}
		sb.WriteString("\n")
	}

	if len(m.Jira) > 0 {
		sb.WriteString("## Jira\n\n")
		for _, j := range m.Jira {
			sb.WriteString(fmt.Sprintf("- %s\n", j))
		}
		sb.WriteString("\n")
	}

	if m.Slack != "" {
		sb.WriteString(fmt.Sprintf("## Slack\n\n%s\n\n", m.Slack))
	}

	if m.Notes != "" {
		sb.WriteString("## Notes\n\n")
		sb.WriteString(strings.TrimSpace(m.Notes))
		sb.WriteString("\n\n")
	}

	sb.WriteString(fmt.Sprintf("---\n_Created: %s_\n", m.CreatedAt.Format("2006-01-02 15:04:05")))

	return sb.String()
}
