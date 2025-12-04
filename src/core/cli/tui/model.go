package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kuchuk-borom-debbarma/GitGrove/core/core/internal/grove"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/core/internal/util/grove"
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
				if err := grove.Initialize(path); err != nil {
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
				if m.choices[m.cursor] == "Init GitGrove" {
					m.inputtingPath = true
					// Update default value to current CWD just in case it changed or wasn't set right
					cwd, _ := os.Getwd()
					m.textInput.SetValue(cwd)
					return m, nil
				} else if m.choices[m.cursor] == "Quit" {
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

	s := ""

	if !m.isGroveRepo {
		if m.inputtingPath {
			s += "Enter path to initialize GitGrove:\n\n"
			s += m.textInput.View()
			s += "\n\n(esc to cancel, enter to confirm)\n"
			if m.err != nil {
				s += fmt.Sprintf("\nError: %v\n", m.err)
			}
		} else {
			s += "Not a GitGrove Repository.\n\n"

			for i, choice := range m.choices {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s += fmt.Sprintf("%s %s\n", cursor, choice)
			}

			if m.err != nil {
				s += fmt.Sprintf("\nError: %v\n", m.err)
			}
		}
	} else {
		s += "Welcome to GitGrove!\n"
		s += "Current Repository Info:\n"
		s += "  Path: " + m.repoInfo + "\n" // Placeholder
		s += "\nPress 'q' to quit.\n"
	}

	return s
}
