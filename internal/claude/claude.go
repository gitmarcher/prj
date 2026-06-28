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

// WriteSkill installs all project slash commands into .claude/commands/.
func WriteSkill(workspaceDir string) error {
	dir := filepath.Join(workspaceDir, ".claude", "commands")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create .claude/commands: %w", err)
	}
	skills := map[string]string{
		"project-analyzer.md": skillAnalyzer,
		"project-reviewer.md": skillReviewer,
	}
	for filename, content := range skills {
		dest := filepath.Join(dir, filename)
		if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
			return fmt.Errorf("cannot write %s: %w", filename, err)
		}
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

	sb.WriteString("## Engineering Principles\n\n")
	sb.WriteString("- Prioritize correctness, simplicity, robustness, maintainability, and scalability over minimizing implementation effort.\n")
	sb.WriteString("- Do not optimize for the smallest diff or the fewest files changed.\n")
	sb.WriteString("- Prefer the solution that best fits the project's long-term architecture, even if it requires more work.\n")
	sb.WriteString("- Use `requirements.md`, `project.md`, and `decisions.md` as the source of truth when making implementation decisions.\n")
	sb.WriteString("- If a proposed change conflicts with an existing architectural decision, stop and explain the conflict instead of silently proceeding.\n")
	sb.WriteString("- When multiple solutions are viable, briefly explain the trade-offs and recommend the one that best supports the long-term health of the system.\n\n")

	sb.WriteString("## Getting Started\n\n")
	sb.WriteString("Before beginning any work in this project:\n\n")
	sb.WriteString("1. Check whether `.prj/project.md` exists.\n")
	sb.WriteString("2. If it does **not** exist, say:\n\n")
	sb.WriteString("   > \"I noticed this project has not been analyzed yet. Would you like me to build an understanding of the project before we begin?\"\n\n")
	sb.WriteString("3. If the user agrees, run `/project-analyzer`.\n\n")
	sb.WriteString("When the implementation is ready for review, run `/project-reviewer`.\n\n")

	sb.WriteString(fmt.Sprintf("---\n_Created: %s_\n", m.CreatedAt.Format("2006-01-02 15:04:05")))

	return sb.String()
}

const skillAnalyzer = `You are running the Project Analyzer skill for a ` + "`prj`" + ` workspace.

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

const skillReviewer = `You are running the Project Reviewer skill for a ` + "`prj`" + ` workspace.

You behave like an experienced Staff Engineer reviewing the entire project — not
just a pull request. You verify correctness, completeness, and architectural
consistency. You do **not** write or fix code.

---

## Step 1 — Load Project Knowledge

Read all available project knowledge before inspecting any code:

- ` + "`.prj/config.yaml`" + ` — project name, repos, branch, intent
- ` + "`.prj/project.md`" + ` — architecture, request flow, impacted components, assumptions
- ` + "`.prj/requirements.md`" + ` — acceptance criteria (skip gracefully if absent)
- ` + "`.prj/decisions.md`" + ` — recorded architectural decisions
- ` + "`.prj/open_questions.md`" + ` — unresolved questions

---

## Step 2 — Inspect All Repositories

Inspect **every** repository listed in ` + "`.prj/config.yaml`" + ` — both primary and secondary.

For each repository read:
- Current branch and working tree state
- Local commits not yet merged
- Diff of all changes relative to the default branch
- Open Pull Requests and CI status (if ` + "`gh`" + ` is available)

Do not skip any repository.

---

## Step 3 — Produce the Review

Work through each responsibility below. Record all findings as you go.

### 3.1 Intent Review

Did the implementation solve the original problem stated in ` + "`.prj/project.md`" + `?
- Was unnecessary scope added?
- Was required functionality omitted?

### 3.2 Requirements Review

For each acceptance criterion in ` + "`.prj/requirements.md`" + `:

Mark each as one of:
- ✅ Complete — with evidence
- ⚠ Partially Complete — with evidence and what is missing
- ❌ Missing — with explanation
- ❓ Unable to Verify — with reason

If ` + "`.prj/requirements.md`" + ` does not exist, derive requirements from the intent and
` + "`.prj/project.md`" + ` and note that no formal requirements file was found.

Every conclusion must cite specific code evidence.

### 3.3 Architectural Decision Review

For each decision in ` + "`.prj/decisions.md`" + `:
- Does the implementation follow this decision?
- If violated: create a **high-priority** finding

### 3.4 Architecture Review

Review the implementation for:
- Incorrect ownership
- Broken request flow
- New coupling or circular dependencies
- Duplicate sources of truth
- Violated assumptions from ` + "`.prj/project.md`" + `
- Incorrect service boundaries

Reason about architecture, not just code.

### 3.5 Repository Boundary Review

**Primary repositories** — expected to contain implementation changes. Review for:
- Correctness and completeness
- Requirements compliance
- Architectural consistency

**Secondary repositories** — not expected to contain changes.

For every secondary repository:
1. Determine whether modifications exist
2. If modifications exist, stop and ask the user:

` + "```" + `
This repository is marked as Secondary: [repo-name]

Changes were detected:
[list the changes]

Were these changes intentional? (yes/no)
` + "```" + `

   - If **yes**: Ask "Why were these changes necessary?" and record the explanation in the review artifact.
   - If **no**: Flag as unintended modification. Treat as a warning; escalate to critical if the changes introduce architectural or behavioral differences.

### 3.6 Cross-Repository Review

Review the project as a whole for incomplete implementation across boundaries:
- API changed → consumers updated?
- Events changed → publishers and consumers consistent?
- Schema changed → migrations added?
- Configuration changed everywhere it is needed?
- Tests added for cross-service contracts?
- Documentation updated?

### 3.7 Missing Work

Search for:
- ` + "`TODO`" + ` / ` + "`FIXME`" + ` / ` + "`HACK`" + ` comments
- Stub or placeholder implementations
- Missing error handlers
- Missing tests (unit, integration, contract)
- Missing input validation
- Missing database migrations
- Missing configuration entries
- Missing event consumers or publishers
- Missing API documentation

### 3.8 Risk Analysis

Identify edge cases, rollback risks, deployment risks, compatibility risks, and
failure scenarios. For each risk explain why it matters and its severity.

### 3.9 Documentation Drift

Determine whether project knowledge should be updated as a result of the
implementation. Examples:
- Should ` + "`.prj/project.md`" + ` be updated to reflect what was actually built?
- Should a new architectural decision be recorded in ` + "`.prj/decisions.md`" + `?
- Should an open question now be resolved?
- Should ` + "`.prj/requirements.md`" + ` be updated?

Do **not** modify these files automatically. Recommend updates only.

### 3.10 Engineering Review

Review for:
- Test coverage and quality
- Error handling completeness
- Logging at appropriate levels
- Security concerns (input validation, auth, secrets, injection)
- Performance concerns (N+1 queries, unbounded loops, missing indexes)
- Obvious code smells

If the ` + "`/no-mistakes`" + ` or ` + "`/code-review`" + ` skill is available, defer to it for
this phase rather than duplicating its work.

---

## Step 4 — Write the Review Artifact

Determine the current date and time. Create the file:

` + "```" + `
.prj/reviews/review-YYYYMMDD-HHMM.md
` + "```" + `

Never overwrite a previous review. Every review is an immutable record.

Use this structure:

` + "```" + `markdown
# Review — [Project Name] — [YYYY-MM-DD HH:MM]

## Metadata
- **Project:** [name]
- **Branch:** [branch]
- **Reviewer:** Project Reviewer skill
- **Date:** [date]
- **Repositories inspected:** [list]

## Overall Status
[One of: ✅ Ready / ⚠ Needs Work / ❌ Blocked]
[One-paragraph summary]

## Summary
[3–5 bullet points covering the most important findings]

## Requirements Review
[Table or list — one row per criterion with status and evidence]

## Architectural Review
[Findings from intent, decisions, architecture, and boundary reviews]

## Repository Boundary Review
[Primary: summary per repo]
[Secondary: any changes found, user responses, flag status]

## Cross-Repository Review
[Findings across repo boundaries]

## Engineering Review
[Tests, error handling, security, performance, code quality]

## Risk Analysis
[Each risk with severity and explanation]

## Documentation Recommendations
[What should be updated and why — no automatic changes]

## Interactive Resolutions
[Findings the user explained or rejected during the session, with their reasoning]

## Next Actions
[Ordered by priority — written so the Project Implementer skill can consume them directly]

Priority 1: [most critical]
Priority 2: ...
` + "```" + `

---

## Step 5 — Interactive Review Session

After writing the artifact, present each finding one at a time:

` + "```" + `
Finding #N — [title]

[Description and evidence]

Actions:
1. Accept finding
2. Reject finding
3. Explain why this is intentional
` + "```" + `

- **Accept:** Mark as confirmed in the artifact.
- **Reject:** Ask for a brief reason; mark as rejected with the reason.
- **Explain:** Prompt "Why is this intentional? > " and record the explanation
  in the **Interactive Resolutions** section of the artifact.
  Do **not** record these explanations in ` + "`.prj/decisions.md`" + ` — review
  explanations are separate from permanent architectural decisions.

Update the artifact after the session is complete.

---

## Constraints

- Never fabricate evidence — every finding must cite specific code or files
- Never assume implementation details without reading the code
- Always inspect every repository listed in the project config
- Primary repositories are expected to change; secondary repositories should not
- Never modify project knowledge files automatically during review
- Every review produces an immutable artifact in ` + "`.prj/reviews/`" + `
- Prioritise correctness, completeness, and architectural consistency over style
- Distinguish permanent architectural decisions from temporary review explanations
`
