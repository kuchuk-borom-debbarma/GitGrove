package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	preparemerge "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/prepare-merge"
	registerrepo "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/register-repo"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
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

// AppState definitions
type AppState int

const (
	StateInit AppState = iota
	StateIdle
	StateInputPath
	StateInputAtomic
	StateRepoSelection
)

type Model struct {
	state        AppState
	repoInfo     string
	err          error
	quitting     bool
	cursor       int
	choices      []string
	repoChoices  []string // List of available repos for selection
	repoCursor   int
	path         string
	textInput    textinput.Model
	descriptions map[string]string
}

func InitialModel() Model {
	cwd, _ := os.Getwd()

	initialState := StateInit
	if err := groveUtil.IsGroveInitialized(cwd); err != nil {
		// IsGroveInitialized returns error if already initialized
		initialState = StateIdle
	} else {
		// No error means not initialized
		initialState = StateInit
	}

	ti := textinput.New()
	ti.Placeholder = "Path to initialize (default: current directory)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	ti.SetValue(cwd)

	// Main menu choices for initialized repo
	mainChoices := []string{"Register Repo (Placeholder)", "Prepare Merge", "Quit"}
	descriptions := make(map[string]string)

	if initialState == StateInit {
		mainChoices = []string{"Init GitGrove", "Quit"}
		descriptions["Init GitGrove"] = initialize.Description()
	} else {
		// Populate descriptions for initialized state
		// Note: "Register Repo (Placeholder)" might not have a direct package link yet if we are not importing it,
		// but we can import registerrepo for Description().
		// We'll need to add imports to model.go
		descriptions["Register Repo (Placeholder)"] = registerrepo.Description()
		// Actually use package description if imported
		// descriptions["Register Repo (Placeholder)"] = registerrepo.Description()
		descriptions["Prepare Merge"] = preparemerge.Description()
	}
	descriptions["Quit"] = "Exit the application."

	return Model{
		state:        initialState,
		choices:      mainChoices,
		textInput:    ti,
		path:         cwd,
		descriptions: descriptions,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state != StateInputPath && m.state != StateInputAtomic {
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	switch m.state {
	case StateInit:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				switch m.choices[m.cursor] {
				case "Init GitGrove":
					m.state = StateInputPath
					cwd, _ := os.Getwd()
					m.textInput.SetValue(cwd)
					return m, nil
				case "Quit":
					m.quitting = true
					return m, tea.Quit
				}
			case "down", "j":
				m.cursor++
				if m.cursor >= len(m.choices) {
					m.cursor = 0
				}
			case "up", "k":
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.choices) - 1
				}
			}
		}

	case StateInputPath:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				path := m.textInput.Value()
				if path == "" {
					path, _ = os.Getwd()
				}
				m.path = path
				m.state = StateInputAtomic
				return m, nil
			case tea.KeyEsc:
				// Go back to Init menu
				m.state = StateInit
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case StateInputAtomic:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				if err := initialize.Initialize(m.path, true); err != nil {
					m.err = err
				} else {
					m.repoInfo = "GitGrove Initialized at " + m.path
					m.state = StateIdle
					m.choices = []string{"Register Repo (Placeholder)", "Prepare Merge", "Quit"}
					m.cursor = 0
				}
				return m, nil
			case "n", "N":
				if err := initialize.Initialize(m.path, false); err != nil {
					m.err = err
				} else {
					m.repoInfo = "GitGrove Initialized at " + m.path
					m.state = StateIdle
					m.choices = []string{"Register Repo (Placeholder)", "Prepare Merge", "Quit"}
					m.cursor = 0
				}
				return m, nil
			case "esc":
				m.state = StateInputPath
				return m, nil
			}
		}

	case StateIdle:
		// Main Menu for authorized repo
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				switch m.choices[m.cursor] {
				case "Register Repo (Placeholder)":
					// Placeholder
				case "Prepare Merge":
					// Check local context
					branch, err := gitUtil.CurrentBranch(m.path)
					if err != nil {
						m.err = err
						return m, nil
					}

					// If orphan -> prepare merge immediately
					if len(branch) > 3 && branch[:3] == "gg/" {
						// Extract repo name? Actually preparemerge handles detection too if we pass empty repoName?
						// Wait, PrepareMerge(cwd, repoName)
						// If we are in orphan branch, we call PrepareMerge(cwd, "").
						if err := preparemerge.PrepareMerge(m.path, ""); err != nil {
							m.err = err
						} else {
							m.repoInfo = "Success: Prepare-merge branch created from orphan branch"
							m.state = StateIdle
						}
						return m, nil
					}

					// If trunk -> go to RepoSelection
					config, err := groveUtil.LoadConfig(m.path)
					if err != nil {
						m.err = err
						return m, nil
					}
					var repos []string
					for name := range config.Repositories {
						repos = append(repos, name)
					}
					m.repoChoices = repos
					m.repoCursor = 0
					m.state = StateRepoSelection
					if len(repos) == 0 {
						// Maybe show error or empty state?
						m.err = fmt.Errorf("no repositories found")
					}

				case "Quit":
					m.quitting = true
					return m, tea.Quit
				}
			case "down", "j":
				m.cursor++
				if m.cursor >= len(m.choices) {
					m.cursor = 0
				}
			case "up", "k":
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.choices) - 1
				}
			}
		}

	case StateRepoSelection:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.state = StateIdle
			case "down", "j":
				m.repoCursor++
				if m.repoCursor >= len(m.repoChoices) {
					m.repoCursor = 0
				}
			case "up", "k":
				m.repoCursor--
				if m.repoCursor < 0 {
					m.repoCursor = len(m.repoChoices) - 1
				}
			case "enter":
				if len(m.repoChoices) > 0 {
					repoName := m.repoChoices[m.repoCursor]
					// Execute Prepare Merge
					if err := preparemerge.PrepareMerge(m.path, repoName); err != nil {
						m.err = err
					} else {
						m.repoInfo = fmt.Sprintf("Success: Prepare-merge branch created for %s", repoName)
						m.state = StateIdle
					}
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
		s += "\n\n" + infoStyle.Render("(esc to cancel, enter to confirm)") + "\n"
		if m.err != nil {
			s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
		}

	case StateInputAtomic:
		s += "Automatically append [RepoName] to commit messages? [y/N]:"

	case StateIdle:
		s += successStyle.Render("Welcome to GitGrove!") + "\n\n"
		s += "Current Repository Info:\n"
		s += "  Path: " + m.repoInfo + "\n" // Placeholder
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
