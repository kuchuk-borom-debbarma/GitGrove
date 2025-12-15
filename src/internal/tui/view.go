package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
)

func (m Model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	var s string

	// Header
	header := titleBorderStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("GitGrove"),
			infoStyle.Render("v1.1.1 ("+m.buildTime+")"),
		),
	)

	switch m.state {
	case StateInit:
		s += "Not a GitGrove Repository.\n\n"

		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				s += selectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
			}
		}

		if m.err != nil {
			s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
		}

	case StateInputPath:
		s += "Enter path to initialize GitGrove:\n\n"
		s += inputStyle.Render(m.textInput.View())
		s += "\n"

		// Render suggestions
		if len(m.suggestions) > 0 {
			s += "\nSuggestions:\n"
			start := 0
			limit := 5
			if m.suggestionCursor >= limit {
				start = m.suggestionCursor - limit + 1
			}
			end := start + limit
			if end > len(m.suggestions) {
				end = len(m.suggestions)
			}

			for i := start; i < end; i++ {
				sugg := m.suggestions[i]
				cursor := " "
				if i == m.suggestionCursor {
					cursor = ">"
					s += selectedItemStyle.Render(fmt.Sprintf("%s 📁 %s", cursor, sugg)) + "\n"
				} else {
					s += itemStyle.Render(fmt.Sprintf("%s 📁 %s", cursor, sugg)) + "\n"
				}
			}
		}

		s += "\n" + infoStyle.Render("(esc to cancel, tab to autocomplete, enter to confirm)") + "\n"
		if m.err != nil {
			s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
		}

	case StateOpenRepoPath:
		s += "Open Existing GitGrove Repository:\n\n"
		s += inputStyle.Render(m.textInput.View())
		s += "\n"

		// Render suggestions (reuse logic)
		if len(m.suggestions) > 0 {
			s += "\nSuggestions:\n"

			start := 0
			limit := 5
			if m.suggestionCursor >= limit {
				start = m.suggestionCursor - limit + 1
			}
			end := start + limit
			if end > len(m.suggestions) {
				end = len(m.suggestions)
			}

			for i := start; i < end; i++ {
				sugg := m.suggestions[i]
				cursor := " "
				if i == m.suggestionCursor {
					cursor = ">"
					s += selectedItemStyle.Render(fmt.Sprintf("%s 📁 %s", cursor, sugg)) + "\n"
				} else {
					s += itemStyle.Render(fmt.Sprintf("%s 📁 %s", cursor, sugg)) + "\n"
				}
			}
		}

		s += "\n" + infoStyle.Render("(esc to cancel, tab to autocomplete, enter to confirm)") + "\n"
		if m.err != nil {
			s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
		}

	case StateInputAtomic:
		s += "Automatically append [RepoName] to commit messages? [y/N]:"
		if m.err != nil {
			s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
		}

	case StateIdle:
		s += successStyle.Render("Welcome to GitGrove!") + "\n\n"
		s += infoStyle.Render("Current Context:") + "\n"
		s += "  " + m.repoInfo + "\n"
		s += "\n"

		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				s += selectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
			}
		}

		if m.err != nil {
			s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
		}

		s += "\n" + infoStyle.Render("Press 'q' to quit.") + "\n"

	case StateRepoSelection:
		s += "Select Repository to Prepare Merge:\n\n"
		if len(m.repoChoices) == 0 {
			s += errorStyle.Render("No repositories found in configuration.") + "\n"
		} else {
			for i, choice := range m.repoChoices {
				cursor := " "
				if m.repoCursor == i {
					cursor = ">"
					s += selectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
				} else {
					s += itemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
				}
			}
		}
		s += "\n" + infoStyle.Render("(esc to cancel, enter to select)") + "\n"
		if m.err != nil {
			s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
		}

	case StateRegisterRepoName:
		s += "Register New Repository\n"
		s += "Enter Repository Name (e.g., service-a):\n\n"
		s += inputStyle.Render(m.textInput.View())
		s += "\n\n" + infoStyle.Render("(esc to cancel, enter to next)") + "\n"

	case StateRegisterRepoPath:
		s += "Register New Repository\n"
		s += "Enter Path for " + m.registerName + ":\n\n"
		s += inputStyle.Render(m.textInput.View())
		s += "\n"

		// Render suggestions
		if len(m.suggestions) > 0 {
			s += "\nSuggestions:\n"
			start := 0
			limit := 5
			if m.suggestionCursor >= limit {
				start = m.suggestionCursor - limit + 1
			}
			end := start + limit
			if end > len(m.suggestions) {
				end = len(m.suggestions)
			}

			for i := start; i < end; i++ {
				sugg := m.suggestions[i]
				cursor := " "
				if i == m.suggestionCursor {
					cursor = ">"
					s += selectedItemStyle.Render(fmt.Sprintf("%s 📁 %s", cursor, sugg)) + "\n"
				} else {
					s += itemStyle.Render(fmt.Sprintf("%s 📁 %s", cursor, sugg)) + "\n"
				}
			}
		}

		s += "\n" + infoStyle.Render("(esc to cancel, tab to autocomplete, enter to confirm)") + "\n"
		if m.err != nil {
			s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
		}

	case StateViewRepos:
		s += "Registered Repositories:\n\n"
		if len(m.repoChoices) == 0 {
			s += errorStyle.Render("No repositories found.") + "\n"
		} else {
			for i, choice := range m.repoChoices {
				cursor := " "
				if m.repoCursor == i {
					cursor = ">"
					s += selectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
				} else {
					s += itemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
				}
			}
		}
		s += "\n" + infoStyle.Render("(esc/q/enter to return)") + "\n"

	case StateRepoCheckoutSelection:
		s += "Select Repository to Checkout Branch:\n\n"
		if len(m.repoChoices) == 0 {
			s += errorStyle.Render("No repositories found in configuration.") + "\n"
		} else {
			for i, choice := range m.repoChoices {
				cursor := " "
				if m.repoCursor == i {
					cursor = ">"
					s += selectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
				} else {
					s += itemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
				}
			}
		}
		s += "\n" + infoStyle.Render("(esc to cancel, enter to checkout)") + "\n"
		if m.err != nil {
			s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
		}
	}

	// Dynamic Description Pane
	// Shows description for the currently selected item in main menus
	var description string
	if m.state == StateInit || m.state == StateIdle {
		if m.cursor >= 0 && m.cursor < len(m.choices) {
			selected := m.choices[m.cursor]
			if desc, ok := m.descriptions[selected]; ok {
				description = desc
			}
		}
	}

	if description != "" {
		// Render description box
		descBox := titleBorderStyle.Render(
			infoStyle.Render(description),
		)
		return appStyle.Render(header + "\n" + s + "\n" + descBox)
	}

	return appStyle.Render(header + "\n" + s)
}

// Helper to get formatted trunk context info
func getTrunkContextInfo(path string, currentBranch string) string {
	config, err := groveUtil.LoadConfig(path)
	if err != nil {
		return fmt.Sprintf("Trunk: %s (Error loading config: %v)", currentBranch, err)
	}

	var repos []string
	for name := range config.Repositories {
		repos = append(repos, name)
	}
	// Sort for consistent display
	sort.Strings(repos)

	repoName := filepath.Base(path)
	info := fmt.Sprintf("Current Repository: %s\n  Trunk: %s", repoName, currentBranch)

	if len(repos) == 0 {
		return fmt.Sprintf("%s\n  (No registered repositories)", info)
	}
	return fmt.Sprintf("%s\n  Registered Repositories:\n    - %s", info, strings.Join(repos, "\n    - "))
}
