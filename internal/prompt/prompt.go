package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

// Ask prompts for a single line of input. Returns "" if the user enters nothing.
func Ask(label string) (string, error) {
	fmt.Printf("%s: ", label)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// AskMultiline prompts for multiple lines until the user submits a blank line.
func AskMultiline(label string) ([]string, error) {
	fmt.Printf("%s (blank line to finish):\n", label)
	var lines []string
	for {
		fmt.Print("  > ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return lines, err
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			break
		}
		lines = append(lines, trimmed)
	}
	return lines, nil
}

// AskOptional is like Ask but prints "(optional)" in the label.
func AskOptional(label string) (string, error) {
	return Ask(label + " (optional)")
}

// AskOptionalMultiline is like AskMultiline but prints "(optional)" and returns nil if empty.
func AskOptionalMultiline(label string) ([]string, error) {
	fmt.Printf("%s (optional, blank line to finish):\n", label)
	var lines []string
	for {
		fmt.Print("  > ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return lines, err
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			break
		}
		lines = append(lines, trimmed)
	}
	return lines, nil
}
