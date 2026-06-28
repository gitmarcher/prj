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

// WriteSkill installs the project-analyzer slash command into .claude/commands/.
func WriteSkill(workspaceDir string) error {
	dir := filepath.Join(workspaceDir, ".claude", "commands")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create .claude/commands: %w", err)
	}
	dest := filepath.Join(dir, "project-analyzer.md")
	if err := os.WriteFile(dest, []byte(skillContent), 0o644); err != nil {
		return fmt.Errorf("cannot write project-analyzer skill: %w", err)
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

	sb.WriteString("## Getting Started\n\n")
	sb.WriteString("Before beginning any work in this project:\n\n")
	sb.WriteString("1. Check whether `.prj/project.md` exists.\n")
	sb.WriteString("2. If it does **not** exist, say:\n\n")
	sb.WriteString("   > \"I noticed this project has not been analyzed yet. Would you like me to build an understanding of the project before we begin?\"\n\n")
	sb.WriteString("3. If the user agrees, run `/project-analyzer`.\n\n")

	sb.WriteString(fmt.Sprintf("---\n_Created: %s_\n", m.CreatedAt.Format("2006-01-02 15:04:05")))

	return sb.String()
}

const skillContent = `You are running the Project Analyzer skill for a ` + "`prj`" + ` workspace.

## Goal

Build and maintain long-term knowledge for this project by analyzing all
repositories and writing structured knowledge files into ` + "`.prj/`" + `.

If ` + "`.prj/project.md`" + ` already exists, read it first and **improve rather than
replace** — merge new discoveries with existing knowledge.

---

## Step 1 — Load Project Context

Read ` + "`.prj/config.yaml`" + ` to get:
- Project name and intent
- Primary and secondary repositories (and feature branch)
- Documentation links, Jira tickets, Slack channel, notes

---

## Step 2 — Analyze Repositories

For **each primary repository**:
- Understand the service's purpose and ownership boundaries
- Identify relevant packages, modules, and APIs
- Find request handlers, event publishers/consumers, database schemas
- Trace the request flow relevant to this project's intent

For **each secondary repository**:
- Identify APIs or events that primary repos depend on
- Note ownership and interface boundaries

Do not summarize things unrelated to the project intent.

---

## Step 3 — Build Understanding

Construct a clear understanding of:

1. **Architecture** — which services are involved, what they own, how they relate
2. **Request Flow** — trace the path relevant to this feature, entry to persistence
3. **Impacted Components** — repos, APIs, events, queues, databases, key classes/functions
4. **Architectural Assumptions** — source of truth, idempotency, retry ownership,
   consistency guarantees, ordering guarantees

---

## Step 4 — Identify Open Questions

When something **cannot be determined confidently** from the code:
- Do **not** guess or invent an answer
- Record it as an open question with the context you do have

---

## Step 5 — Interactive Clarification

For each open question, ask the user one at a time:

` + "```" + `
Open Question (N/M)

[Question]

Current understanding:
[What the code suggests, without guessing]

Your answer:
>
(Press Enter to skip)
` + "```" + `

**If the user answers:**
- Follow up: "Why was this decision made? > "
- Append to ` + "`.prj/decisions.md`" + `:
  ` + "```" + `
  ## [Decision title]
  **Decision:** [answer]
  **Reason:** [why]
  **Alternatives:** [if mentioned]
  **Date:** [today]
  ` + "```" + `
- Remove this question from the open questions list

**If the user skips:**
- Keep the question unresolved; never invent an answer

---

## Step 6 — Write Knowledge Files

### ` + "`.prj/project.md`" + `

Write a comprehensive document with these sections:

` + "```" + `markdown
# [Project Name]

## Intent
[What this feature is trying to accomplish]

## Architecture
[Services involved, ownership boundaries, how they relate]

## Request Flow
[Step-by-step path relevant to this feature]

## Impacted Components
### Repositories
### APIs
### Events / Queues
### Databases
### Key Classes / Functions

## Assumptions
[Source of truth, idempotency, retry ownership, consistency guarantees]

## Open Questions
[Summary of unresolved questions — full detail in open_questions.md]

## Documentation
[Links and references from .prj/config.yaml]
` + "```" + `

### ` + "`.prj/decisions.md`" + `

**Append** new decisions — never overwrite previous ones.

If the file does not exist, create it with a header:
` + "```" + `markdown
# Architectural Decisions
` + "```" + `

### ` + "`.prj/open_questions.md`" + `

Maintain every unresolved question. Format:
` + "```" + `markdown
## [Question]
**Context:** [what the code suggests]
**Status:** Unresolved
` + "```" + `

Remove questions that were answered during Step 5.

---

## Constraints

- Never fabricate architectural knowledge
- Preserve all existing user-written content
- Merge new discoveries with existing knowledge; do not regenerate from scratch
- If unsure about something, create an open question instead of guessing
- The output must enable any engineer — or future Claude session — to immediately
  understand this project's intent, architecture, decisions, and remaining unknowns
`

