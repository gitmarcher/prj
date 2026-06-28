# prj

Bootstrap multi-repository engineering workspaces.

`prj` creates a project directory, clones repositories, creates feature branches,
collects project context, and installs Claude Code skills for AI-assisted development.

---

## Installation

```bash
go install github.com/gitmarcher/prj@latest
```

---

## Configuration

On first run, `prj` creates `~/.config/prj/config.yaml`. Add your sources:

```yaml
workspace: ~/Projects

sources:
  carousell:
    base: git@github.com:carousell
```

| Field | Default | Description |
|---|---|---|
| `workspace` | `~/Projects` | Root directory for all project workspaces |
| `sources` | — | Named repository hosts |
| `knowledge_age_days` | `7` | Days before `project.md` is considered stale |

---

## Commands

### `prj create`

```bash
prj create <project-name> \
  --src=<source> \
  --p <repo1> <repo2> ... \
  --s <repo1> <repo2> ... \
  --b <branch-name>
```

Creates a workspace, clones repositories, creates the feature branch on primary
repos, collects project context interactively, and generates `CLAUDE.md` and
`.claude/commands/` skills.

### `prj repo add`

```bash
prj repo add --p <repo>
prj repo add --s <repo>
```

Adds a repository to an existing project. Only one type (`--p` or `--s`) per
invocation. Updates `.prj/config.yaml` and regenerates `CLAUDE.md`. Marks
project knowledge as stale until `/project-analyzer` is re-run.

### `prj status`

```bash
prj status
```

Shows a consolidated dashboard: git state, ahead/behind, PR status, CI status,
and project knowledge health across all repositories.

---

## Workspace Layout

```
prj-<name>/
├── CLAUDE.md                        auto-generated AI context file
├── .prj/
│   ├── config.yaml                  project metadata (source of truth)
│   ├── project.md                   written by /project-analyzer
│   ├── decisions.md                 written by /project-analyzer + /project-reviewer
│   ├── open_questions.md            written by /project-analyzer
│   └── reviews/
│       └── review-YYYYMMDD-HHMM.md written by /project-reviewer
├── .claude/
│   └── commands/
│       ├── project-analyzer.md      /project-analyzer skill
│       └── project-reviewer.md      /project-reviewer skill
├── <primary-repo>/                  cloned on feature branch
└── <secondary-repo>/                cloned on default branch
```

---

## Claude Code Skills

Skills are installed automatically into every workspace.

| Command | Purpose |
|---|---|
| `/project-analyzer` | Analyzes all repositories, builds `project.md`, `decisions.md`, `open_questions.md` |
| `/project-reviewer` | Full project review — intent, requirements, architecture, cross-repo consistency. Produces an immutable review artifact in `.prj/reviews/`. |

---

## Example

```bash
prj create checkout-v2 \
  --src=carousell \
  --p marketplace-api checkout-service payment-service \
  --s pricing-service analytics \
  --b feature/checkout-v2
```
