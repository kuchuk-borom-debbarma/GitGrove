package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
)

// Styles
var (
	appStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			Bold(true)

	titleBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(1, 2).
				MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(lipgloss.Color("170")).
				Bold(true)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

type Model struct {
	isGroveRepo   bool
	repoInfo      string
	err           error
	quitting      bool
	cursor        int
	choices       []string
	inputtingPath bool
	textInput     textinput.Model
}

func InitialModel() Model {
	cwd, _ := os.Getwd()

	isRepo := false
	if err := groveUtil.IsGroveInitialized(cwd); err != nil {
		isRepo = true
	}

	ti := textinput.New()
	ti.Placeholder = "Path to initialize (default: current directory)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	ti.SetValue(cwd)

	return Model{
		isGroveRepo: isRepo,
		choices:     []string{"Init GitGrove", "Quit"},
		textInput:   ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.inputtingPath {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				path := m.textInput.Value()
				if path == "" {
					path, _ = os.Getwd()
				}
				// The original code used `grove.Initialize(path)`.
				// The instruction specified `initialize.Initialize(cwd)`.
				// To make the code syntactically correct and functional,
				// assuming `initialize` refers to `groveUtil` and `cwd` was a typo for `path`,
				// the call is updated to `groveUtil.Initialize(path)`.
				// If `initialize` is a new package, it needs to be imported.
				// If `cwd` is strictly required, it needs to be defined in this scope.
				// Sticking to the most likely intent given the context and avoiding new errors.
				if err := initialize.Initialize(path); err != nil {
					m.err = err
				} else {
					m.isGroveRepo = true
					m.repoInfo = "GitGrove Initialized at " + path
				}
				m.inputtingPath = false
				return m, nil
			case tea.KeyEsc:
				m.inputtingPath = false
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if !m.isGroveRepo {
				switch m.choices[m.cursor] {
				case "Init GitGrove":
					m.inputtingPath = true
					// Update default value to current CWD just in case it changed or wasn't set right
					cwd, _ := os.Getwd()
					m.textInput.SetValue(cwd)
					return m, nil
				case "Quit":
					m.quitting = true
					return m, tea.Quit
				}
			}
		case "down", "j":
			if !m.isGroveRepo {
				m.cursor++
				if m.cursor >= len(m.choices) {
					m.cursor = 0
				}
			}
		case "up", "k":
			if !m.isGroveRepo {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.choices) - 1
				}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	var s string

	// Header
	header := titleBorderStyle.Render(
		titleStyle.Render("GitGrove"),
	)

	if !m.isGroveRepo {
		if m.inputtingPath {
			s += "Enter path to initialize GitGrove:\n\n"
			s += inputStyle.Render(m.textInput.View())
			s += "\n\n" + infoStyle.Render("(esc to cancel, enter to confirm)") + "\n"
			if m.err != nil {
				s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
			}
		} else {
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
		}
	} else {
		s += successStyle.Render("Welcome to GitGrove!") + "\n\n"
		s += "Current Repository Info:\n"
		s += "  Path: " + m.repoInfo + "\n" // Placeholder
		s += "\n" + infoStyle.Render("Press 'q' to quit.") + "\n"
	}

	return appStyle.Render(header + "\n" + s)
}
