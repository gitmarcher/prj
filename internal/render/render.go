package render

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/gitmarcher/prj/internal/gh"
	"github.com/gitmarcher/prj/internal/status"
)

// ANSI color vars — zeroed when output is not a TTY or NO_COLOR is set.
var (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	white  = "\033[97m"
	gray   = "\033[90m"
)

const lineWidth = 72

func init() {
	if os.Getenv("NO_COLOR") != "" || !isTerminal() {
		reset = ""
		bold = ""
		dim = ""
		green = ""
		yellow = ""
		red = ""
		cyan = ""
		white = ""
		gray = ""
	}
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// Dashboard writes the full project status dashboard to stdout.
func Dashboard(ps *status.ProjectStatus, ghAvailable bool) {
	sep := gray + strings.Repeat("━", lineWidth) + reset

	// ── Header ──────────────────────────────────────────────────────────
	fmt.Println(sep)
	left := fmt.Sprintf("  %sprj%s  %s%s%s", bold, reset, white+bold, ps.Meta.Name, reset)
	right := cyan + ps.Meta.Branch + reset
	// visible length: "  prj  <name>" and "<branch>"
	visLeft := 7 + len(ps.Meta.Name)
	visRight := len(ps.Meta.Branch)
	pad := lineWidth - visLeft - visRight - 2
	if pad < 1 {
		pad = 1
	}
	fmt.Printf("%s%s%s\n", left, strings.Repeat(" ", pad), right)
	fmt.Println(sep)
	fmt.Println()

	// ── Repositories ────────────────────────────────────────────────────
	var primary, secondary []*status.RepoStatus
	for _, rs := range ps.Repos {
		if rs.Role == "primary" {
			primary = append(primary, rs)
		} else {
			secondary = append(secondary, rs)
		}
	}

	if len(primary) > 0 {
		fmt.Printf("  %s%sPRIMARY%s\n\n", bold, white, reset)
		for _, rs := range primary {
			printRepo(rs, ghAvailable)
		}
	}
	if len(secondary) > 0 {
		fmt.Printf("  %s%sSECONDARY%s\n\n", bold, white, reset)
		for _, rs := range secondary {
			printRepo(rs, ghAvailable)
		}
	}

	// ── Summary ─────────────────────────────────────────────────────────
	fmt.Println(sep)
	fmt.Println()
	printSummary(ps, ghAvailable)
	fmt.Println()

	// ── Knowledge ───────────────────────────────────────────────────────
	fmt.Printf("  %s%sKNOWLEDGE%s\n\n", bold, white, reset)
	printKnowledge(ps)
	fmt.Println()
	fmt.Println(sep)
}

// ── repo block ──────────────────────────────────────────────────────────

func printRepo(rs *status.RepoStatus, ghAvailable bool) {
	icon, iconColor := repoIcon(rs)
	fmt.Printf("  %s%s%s  %s%s%s\n", iconColor, icon, reset, bold, rs.Name, reset)

	if rs.Err != nil {
		fmt.Printf("       %s%s%s\n\n", red, rs.Err.Error(), reset)
		return
	}

	// branch line
	branchParts := []string{rs.Branch}
	if rs.DirtyCount == 0 {
		branchParts = append(branchParts, green+"✓ clean"+reset)
	} else {
		branchParts = append(branchParts, fmt.Sprintf("%s✗ %d change%s%s", red, rs.DirtyCount, plural(rs.DirtyCount), reset))
	}
	if !rs.NoUpstream {
		branchParts = append(branchParts, formatAheadBehind(rs.Ahead, rs.Behind))
	} else {
		branchParts = append(branchParts, gray+"no upstream"+reset)
	}
	fmt.Printf("     %sbranch%s  %s\n", dim, reset, strings.Join(branchParts, "  "))

	// PR line
	if !ghAvailable {
		// skip silently; warn once in summary
	} else if rs.PR == nil {
		fmt.Printf("     %sPR%s      %snone%s\n", dim, reset, gray, reset)
	} else {
		printPR(rs.PR)
		if rs.CI != nil {
			printCI(rs.CI)
		}
	}

	fmt.Println()
}

func printPR(pr *gh.PR) {
	state, stateColor := prState(pr)
	title := truncate(pr.Title, 48)
	fmt.Printf("     %sPR%s      #%d  %s%s%s  ·  %s\n", dim, reset, pr.Number, stateColor, state, reset, title)
	fmt.Printf("             %s%s%s\n", gray, pr.URL, reset)
}

func printCI(run *gh.Run) {
	label, color := ciLabel(run)
	workflow := run.WorkflowName
	if workflow == "" {
		workflow = "workflow"
	}
	fmt.Printf("     %sCI%s      %s%s%s  ·  %s\n", dim, reset, color, label, reset, workflow)
}

// ── summary ─────────────────────────────────────────────────────────────

func printSummary(ps *status.ProjectStatus, ghAvailable bool) {
	total := len(ps.Repos)
	dirty, prOpen, prDraft, prMerged, ciFailing := 0, 0, 0, 0, 0

	for _, rs := range ps.Repos {
		if rs.Err != nil {
			continue
		}
		if rs.DirtyCount > 0 {
			dirty++
		}
		if rs.PR != nil {
			switch rs.PR.State {
			case "OPEN":
				if rs.PR.IsDraft {
					prDraft++
				} else {
					prOpen++
				}
			case "MERGED":
				prMerged++
			}
		}
		if rs.CI != nil && rs.CI.Status == "completed" && rs.CI.Conclusion == "failure" {
			ciFailing++
		}
	}

	parts := []string{fmt.Sprintf("%d repo%s", total, plural(total))}
	if dirty > 0 {
		parts = append(parts, fmt.Sprintf("%s%d with changes%s", yellow, dirty, reset))
	} else {
		parts = append(parts, fmt.Sprintf("%s0 with changes%s", dim, reset))
	}
	if ghAvailable {
		if prOpen > 0 {
			parts = append(parts, fmt.Sprintf("%s%d PR open%s", green, prOpen, reset))
		} else {
			parts = append(parts, fmt.Sprintf("%s0 PR open%s", dim, reset))
		}
		if prDraft > 0 {
			parts = append(parts, fmt.Sprintf("%s%d draft%s", gray, prDraft, reset))
		}
		if prMerged > 0 {
			parts = append(parts, fmt.Sprintf("%s%d merged%s", cyan, prMerged, reset))
		}
		if ciFailing > 0 {
			parts = append(parts, fmt.Sprintf("%s%d CI failing%s", red, ciFailing, reset))
		} else {
			parts = append(parts, fmt.Sprintf("%s0 CI failing%s", dim, reset))
		}
	} else {
		parts = append(parts, gray+"gh CLI not found — PR/CI unavailable"+reset)
	}

	fmt.Printf("  %s%sSUMMARY%s  %s\n", bold, white, reset, strings.Join(parts, "  ·  "))
}

// ── knowledge ────────────────────────────────────────────────────────────

func printKnowledge(ps *status.ProjectStatus) {
	for _, kf := range ps.Knowledge {
		if !kf.Exists {
			fmt.Printf("  %s✗%s  %-22s%smissing%s\n", red, reset, kf.Name, gray, reset)
			continue
		}
		// project.md is additionally marked ⚠ when repos were added after it was written
		if kf.Name == "project.md" && len(ps.StaleAdditions) > 0 {
			fmt.Printf("  %s⚠%s  %-22s%slast updated %s  ·  %d repo%s added since last analysis%s\n",
				yellow, reset, kf.Name, yellow, timeAgo(kf.ModTime),
				len(ps.StaleAdditions), plural(len(ps.StaleAdditions)), reset)
			continue
		}
		if kf.IsStale {
			fmt.Printf("  %s~%s  %-22s%slast updated %s  (consider refreshing)%s\n",
				yellow, reset, kf.Name, yellow, timeAgo(kf.ModTime), reset)
		} else {
			fmt.Printf("  %s✓%s  %-22s%slast updated %s%s\n",
				green, reset, kf.Name, dim, timeAgo(kf.ModTime), reset)
		}
	}

	if len(ps.StaleAdditions) > 0 {
		fmt.Println()
		fmt.Printf("  %s%sRepositories added since last analysis%s\n", bold, yellow, reset)
		for _, a := range ps.StaleAdditions {
			fmt.Printf("    %s%-26s%s %s%-10s%s  %s%s%s\n",
				bold, a.Name, reset,
				gray, "("+a.Role+")", reset,
				dim, a.Reason, reset)
		}
		fmt.Printf("\n  %sRun /project-analyzer to refresh the project's understanding.%s\n", gray, reset)
		return
	}

	// If project.md is missing, print an action hint
	for _, kf := range ps.Knowledge {
		if kf.Name == "project.md" && !kf.Exists {
			fmt.Printf("\n  %sProject has not yet been analyzed.%s\n", yellow, reset)
			fmt.Printf("  %sOpen Claude Code and run the Project Analyzer skill.%s\n", gray, reset)
		}
	}
}

// ── helpers ──────────────────────────────────────────────────────────────

func repoIcon(rs *status.RepoStatus) (string, string) {
	if rs.Err != nil {
		return "!", red
	}
	if rs.CI != nil && rs.CI.Status == "completed" && rs.CI.Conclusion == "failure" {
		return "✗", red
	}
	if rs.DirtyCount > 0 {
		return "~", yellow
	}
	if rs.Behind > 0 {
		return "↓", yellow
	}
	return "✓", green
}

func prState(pr *gh.PR) (string, string) {
	if pr.IsDraft {
		return "Draft", gray
	}
	switch pr.State {
	case "OPEN":
		return "Open", green
	case "MERGED":
		return "Merged", cyan
	case "CLOSED":
		return "Closed", red
	default:
		return pr.State, gray
	}
}

func ciLabel(run *gh.Run) (string, string) {
	if run.Status == "in_progress" {
		return "running", yellow
	}
	if run.Status == "queued" {
		return "queued", yellow
	}
	switch run.Conclusion {
	case "success":
		return "✓ passing", green
	case "failure":
		return "✗ failing", red
	case "cancelled":
		return "cancelled", gray
	default:
		return run.Status, gray
	}
}

func formatAheadBehind(ahead, behind int) string {
	var parts []string
	if ahead == 0 {
		parts = append(parts, dim+"↑0"+reset)
	} else {
		parts = append(parts, fmt.Sprintf("↑%d", ahead))
	}
	if behind == 0 {
		parts = append(parts, dim+"↓0"+reset)
	} else {
		parts = append(parts, fmt.Sprintf("%s↓%d%s", yellow, behind, reset))
	}
	return strings.Join(parts, " ")
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	days := int(math.Floor(d.Hours() / 24))
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case days == 1:
		return "1 day ago"
	default:
		return fmt.Sprintf("%d days ago", days)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
