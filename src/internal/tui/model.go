package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	preparemerge "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/prepare-merge"
	registerrepo "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/register-repo"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
)

// Styles
var (
	appStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	titleBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(1, 2).
				MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(lipgloss.Color("#EE6FF8")).
				Bold(true)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4C4C")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FA9A")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))
)

// AppState definitions
type AppState int

const (
	StateInit AppState = iota
	StateIdle
	StateInputPath
	StateInputAtomic
	StateOpenRepoPath
	StateRepoSelection
	StateRegisterRepoName
	StateRegisterRepoPath
	StateViewRepos
	StateRepoCheckoutSelection
)

type Model struct {
	state            AppState
	repoInfo         string
	err              error
	quitting         bool
	cursor           int
	choices          []string
	repoChoices      []string // List of available repos for selection
	repoCursor       int
	path             string
	textInput        textinput.Model
	descriptions     map[string]string
	registerName     string   // Name for the new repo
	isOrphan         bool     // True if in orphan branch
	orphanRepoName   string   // Name of repo if in orphan branch
	trunkBranch      string   // Name of trunk branch if in orphan branch
	suggestions      []string // Autocompletion suggestions
	suggestionCursor int      // Selected suggestion index
	buildTime        string   // Build time of the binary
}

func InitialModel(buildTime string) Model {
	cwd, _ := os.Getwd()

	initialState := StateInit
	var repoInfo string
	var isOrphan bool
	var orphanRepoName, trunkBranch string

	// Check initialization status (returns error if initialized)
	errInit := groveUtil.IsGroveInitialized(cwd)

	if errInit != nil {
		initialState = StateIdle
		// Determine context: Trunk or Orphan?
		currentBranch, err := gitUtil.CurrentBranch(cwd)
		if err == nil {
			// Check for Orphan Pattern: gg/<trunk>/<repoName>
			// We can use the same logic as in grove_util or prepare_merge
			// Or just simple prefix check for TUI display purposes
			if len(currentBranch) > 3 && currentBranch[:3] == "gg/" {
				isOrphan = true
				// Parse: gg/main/serviceA -> trunk: main, repo: serviceA
				// Assumption: trunk doesn't have slashes, or we rely on standard format
				parts := strings.Split(currentBranch, "/")
				if len(parts) >= 3 {
					orphanRepoName = parts[len(parts)-1]
					trunkBranch = strings.Join(parts[1:len(parts)-1], "/")
					repoInfo = fmt.Sprintf("Orphan Branch: %s (Trunk: %s)", orphanRepoName, trunkBranch)
				} else {
					repoInfo = fmt.Sprintf("Orphan Branch: %s", currentBranch)
				}
			} else {
				// Trunk context
				repoInfo = getTrunkContextInfo(cwd, currentBranch)
			}
		} else {
			repoInfo = cwd // Fallback
		}
	} else {
		// Not initialized
		initialState = StateInit
	}

	ti := textinput.New()
	ti.Placeholder = "Path to initialize (default: current directory)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	ti.SetValue(cwd)

	// Main menu choices based on state
	var mainChoices []string
	descriptions := make(map[string]string)

	if initialState == StateInit {
		mainChoices = []string{"Init GitGrove", "Open Repository", "Quit"}
		descriptions["Init GitGrove"] = "Initialize a new GitGrove workspace in the current directory."
		descriptions["Open Repository"] = "Open an existing GitGrove repository located elsewhere."
	} else {
		if isOrphan {
			mainChoices = []string{"Prepare Merge", "Return to Trunk", "Quit"}
			descriptions["Prepare Merge"] = "Prepare the current orphan branch for merging back into the trunk."
			descriptions["Return to Trunk"] = fmt.Sprintf("Checkout the trunk branch (%s) and leave the orphan state.", trunkBranch)
		} else {
			mainChoices = []string{"View Repos", "Register Repo", "Checkout Repo Branch", "Quit"}
			descriptions["View Repos"] = "View a list of all registered repositories in this workspace."
			descriptions["Register Repo"] = "Register a new repository (subdirectory) and create its orphan branch."
			descriptions["Checkout Repo Branch"] = "Switch context to a specific repository's orphan branch."
		}
	}
	descriptions["Quit"] = "Exit the GitGrove application."

	return Model{
		state:            initialState,
		choices:          mainChoices,
		textInput:        ti,
		path:             cwd,
		repoInfo:         repoInfo,
		isOrphan:         isOrphan,
		orphanRepoName:   orphanRepoName,
		trunkBranch:      trunkBranch,
		descriptions:     descriptions,
		suggestionCursor: -1,
		buildTime:        buildTime,
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

		// case tea.PasteMsg:
		// 	var cmd tea.Cmd
		// 	m.textInput, cmd = m.textInput.Update(msg)
		// 	return m, cmd
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
					m.err = nil
					cwd, _ := os.Getwd()
					m.textInput.SetValue(cwd)
					// Initialize suggestions for CWD
					// For Init, we actually want suggestions for THE PATH being typed.
					// Initially it's CWD.
					// But usually Init path suggestions should be relative to where?
					// If input is absolute path, getSuggestions might need adjustment.
					// Current getSuggestions assumes relative to basePath.
					// If m.textInput has absolute path CWD, getSuggestions(cwd, cwd) -> might conform to logic?
					// If input "..." -> dir "." -> search CWD.
					// If input starts with /, it might be treated as absolute.
					// Let's rely on getSuggestions handling or simply show nothing initially if full path is set.
					// BUT for RegisterRepoPath and OpenRepoPath, user starts with empty or relative.

					// User requested: "always show the path suggestion i dont see it until i type any paths"
					// For Init: Input is pre-filled with CWD. Suggestions for siblings of CWD?
					// Or children?
					// If user clears it, they get suggestions.
					// Let's reset suggestions to nil here because fully qualified path is likely valid/done.
					// Or trigger it?
					// Let's stick to user request for the OTHER flows primarily if Init is pre-filled.
					// But user mentioned "init path as well as register path".
					// So let's trigger it.
					m.suggestions = getSuggestions(cwd, m.textInput.Value())
					m.suggestionCursor = -1
					return m, nil
				case "Open Repository":
					m.state = StateOpenRepoPath
					m.err = nil
					cwd, _ := os.Getwd()
					m.textInput.SetValue("") // User starts empty? Or CWD? "Path to existing..." usually implies search.
					// Previous code set value to CWD. User might want to browse from CWD.
					// Let's set to "" to force browsing? Or CWD?
					// If CWD, they see children of CWD?
					// Let's set to "" so they see top level dirs of CWD.
					// Wait, if I set "" and call getSuggestions(cwd, ""), I get children of cwd.
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Path to existing GitGrove repository"
					m.suggestions = getSuggestions(cwd, "")
					m.suggestionCursor = -1
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

	case StateOpenRepoPath:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				// Handle suggestion selection
				if len(m.suggestions) > 0 && m.suggestionCursor >= 0 {
					selected := m.suggestions[m.suggestionCursor]
					if m.textInput.Value() != selected {
						m.textInput.SetValue(selected)
						m.textInput.CursorEnd()
						cwd, _ := os.Getwd()
						// OpenRepoPath uses cwd in Update suggestions loop
						m.suggestions = getSuggestions(cwd, m.textInput.Value())
						m.suggestionCursor = -1
						return m, nil
					}
				}

				path := m.textInput.Value()
				if path == "" {
					path, _ = os.Getwd()
				}

				// Validate if it is a GitGrove repo
				err := groveUtil.IsGroveInitialized(path)
				// IsGroveInitialized returns error if INITIALIZED (which is success for us)
				// It returns nil if NOT initialized (failure for us)

				if err != nil {
					// Check if error message confirms initialization
					if err.Error() == fmt.Sprintf("gitgrove is already initialized in %s", path) ||
						(len(err.Error()) > 30 && err.Error()[:31] == "gitgrove is already initialized") {
						// Success
						m.path = path

						// Get context info
						// Get context info
						currentBranch, _ := gitUtil.CurrentBranch(path)
						if len(currentBranch) > 3 && currentBranch[:3] == "gg/" {
							m.isOrphan = true
							parts := strings.Split(currentBranch, "/")
							if len(parts) >= 3 {
								orphanRepoName := parts[len(parts)-1]
								trunkBranch := strings.Join(parts[1:len(parts)-1], "/")
								m.repoInfo = fmt.Sprintf("Orphan Branch: %s (Trunk: %s)", orphanRepoName, trunkBranch)
								m.orphanRepoName = orphanRepoName
								m.trunkBranch = trunkBranch
							} else {
								m.repoInfo = fmt.Sprintf("Orphan Branch: %s", currentBranch)
							}
							m.choices = []string{"Prepare Merge", "Return to Trunk", "Quit"}
						} else {
							m.isOrphan = false
							m.repoInfo = getTrunkContextInfo(path, currentBranch)
							m.choices = []string{"View Repos", "Register Repo", "Checkout Repo Branch", "Quit"}
						}

						m.state = StateIdle
						m.cursor = 0
						m.err = nil
					} else {
						// Some other error
						m.err = err
					}
				} else {
					// err == nil means not initialized
					m.err = fmt.Errorf("path '%s' is not a GitGrove repository", path)
				}
				return m, nil

			case tea.KeyEsc:
				// Go back to Init menu
				m.state = StateInit
				m.err = nil
				return m, nil

			case tea.KeyTab:
				if len(m.suggestions) > 0 {
					if m.suggestionCursor < 0 {
						m.suggestionCursor = 0
					}
					m.textInput.SetValue(m.suggestions[m.suggestionCursor])
					m.textInput.CursorEnd()
					m.suggestions = nil
					cwd, _ := os.Getwd()
					m.suggestions = getSuggestions(cwd, m.textInput.Value())
					m.suggestionCursor = -1
				}
				return m, nil
			case tea.KeyUp:
				if len(m.suggestions) > 0 {
					m.suggestionCursor--
					if m.suggestionCursor < 0 {
						// Don't wrap to bottom, stay at -1 (input field focus) or 0?
						// Standard dropdown: -1 usually means "no selection" (focus on input).
						// If user presses UP from 0, go to -1?
						// Previous behavior: wrap to bottom.
						// User said "shoots up to the first path" - maybe they meant it wraps to top when at bottom?
						// "select press arrow to select second or other path it shoots up to the first path"
						// This sounds like index 0 is being selected unexpectedly.
						// If cursor wraps, pressing DOWN at bottom goes to top.
						// Pressing DOWN at index 0 should go to index 1.
						m.suggestionCursor = -1 // Stop at input
					}
				}
				return m, nil
			case tea.KeyDown:
				if len(m.suggestions) > 0 {
					m.suggestionCursor++
					if m.suggestionCursor >= len(m.suggestions) {
						// Stop at last item, don't wrap to top
						m.suggestionCursor = len(m.suggestions) - 1
					}
				}
				return m, nil
			}
		}

		m.textInput, cmd = m.textInput.Update(msg)

		// Update suggestions
		cwd, _ := os.Getwd()
		m.suggestions = getSuggestions(cwd, m.textInput.Value())
		m.suggestionCursor = -1
		return m, cmd

	case StateInputPath:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				// If suggestions are active
				if len(m.suggestions) > 0 && m.suggestionCursor >= 0 {
					selected := m.suggestions[m.suggestionCursor]

					// If the input ALREADY matches the suggestion, treat this as a submission (fall through)
					// Or if the user wants to drill down, they used Tab or Enter previously.
					// If they press Enter on "src/", and text input is "src/", we should submit "src/"?
					// Yes, usually.
					// But what if they want to see children of src/?
					// They can type slash? "src/" -> trigger?
					// Should Enter on "src/" (selected) -> set "src/" -> refresh suggestions to children?
					// If suggestions update to children, then input != new selection (children).

					// Improved Logic:
					// 1. If text != selected: Set text = selected. Refresh suggestions.
					// 2. If text == selected: Submit.

					if m.textInput.Value() != selected {
						m.textInput.SetValue(selected)
						m.textInput.CursorEnd()

						// Refresh suggestions (Drill down)
						cwd, _ := os.Getwd()
						m.suggestions = getSuggestions(cwd, m.textInput.Value())
						m.suggestionCursor = -1
						return m, nil
					}
					// If equal, fall through to submit logic
				}

				path := m.textInput.Value()
				if path == "" {
					path, _ = os.Getwd()
				}

				// Check if already initialized BEFORE moving to next state
				if err := groveUtil.IsGroveInitialized(path); err != nil {
					m.err = err
					return m, nil
				}

				m.path = path
				m.state = StateInputAtomic
				return m, nil

				// ... (skipping to View)

			case tea.KeyEsc:
				// Go back to Init menu
				m.state = StateInit
				m.err = nil
				return m, nil
			case tea.KeyTab:
				if len(m.suggestions) > 0 {
					if m.suggestionCursor < 0 {
						m.suggestionCursor = 0
					}
					m.textInput.SetValue(m.suggestions[m.suggestionCursor])
					m.textInput.CursorEnd()
					// Re-fetch suggestions
					cwd, _ := os.Getwd()
					m.suggestions = getSuggestions(cwd, m.textInput.Value())
					m.suggestionCursor = -1
				}
				return m, nil
			case tea.KeyUp:
				if len(m.suggestions) > 0 {
					m.suggestionCursor--
					if m.suggestionCursor < 0 {
						m.suggestionCursor = -1
					}
				}
				return m, nil // Don't process this key in textinput
			case tea.KeyDown:
				if len(m.suggestions) > 0 {
					m.suggestionCursor++
					if m.suggestionCursor >= len(m.suggestions) {
						m.suggestionCursor = len(m.suggestions) - 1
					}
				}
				return m, nil // Don't process this key in textinput
			}
		}

		m.textInput, cmd = m.textInput.Update(msg)

		// Update suggestions
		// For Init path, we assume relative to CWD or absolute?
		// Text input is pre-filled with CWD.
		// getSuggestions is designed for relative paths inside repo root.
		// Use "." (Current Directory) as base for Init.
		cwd, _ := os.Getwd()
		m.suggestions = getSuggestions(cwd, m.textInput.Value())
		m.suggestionCursor = -1 // Reset cursor on new input

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
					m.isOrphan = false
					m.state = StateIdle
					m.choices = []string{"View Repos", "Register Repo", "Checkout Repo Branch", "Quit"}
					m.cursor = 0
				}
				return m, nil
			case "n", "N":
				if err := initialize.Initialize(m.path, false); err != nil {
					m.err = err
				} else {
					m.repoInfo = "GitGrove Initialized at " + m.path
					m.isOrphan = false
					m.state = StateIdle
					m.choices = []string{"View Repos", "Register Repo", "Checkout Repo Branch", "Quit"}
					m.cursor = 0
				}
				return m, nil
			case "esc":
				m.state = StateInputPath
				m.err = nil
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
				case "Register Repo":
					m.state = StateRegisterRepoName
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Enter repository name (e.g., service-a)"
					m.textInput.Focus()
					return m, nil
				case "Prepare Merge":
					// Check local context
					branch, err := gitUtil.CurrentBranch(m.path)
					if err != nil {
						m.err = err
						return m, nil
					}

					// If orphan -> prepare merge immediately
					if len(branch) > 3 && branch[:3] == "gg/" {
						if err := preparemerge.PrepareMerge(m.path, ""); err != nil {
							m.err = err
						} else {
							m.repoInfo = "Success: Prepare-merge branch created from orphan branch"
							m.state = StateIdle
						}
						return m, nil
					}
					// If not orphan, something is weird because this option should only be available in orphan state.
					m.err = fmt.Errorf("prepare merge only available in orphan branches")
					return m, nil

				case "Return to Trunk":
					if m.trunkBranch == "" {
						m.err = fmt.Errorf("unknown trunk branch")
						return m, nil
					}
					if err := gitUtil.Checkout(m.path, m.trunkBranch); err != nil {
						m.err = fmt.Errorf("failed to checkout trunk: %v", err)
					} else {
						// Checked out successfully.
						// Re-evaluate context.
						// Or just assume we are back to trunk.
						m.isOrphan = false
						m.repoInfo = getTrunkContextInfo(m.path, m.trunkBranch)
						m.choices = []string{"View Repos", "Register Repo", "Checkout Repo Branch", "Quit"}
						m.state = StateIdle
					}
					return m, nil

				case "View Repos":
					config, err := groveUtil.LoadConfig(m.path)
					if err != nil {
						m.err = err
						return m, nil
					}
					var repos []string
					for name := range config.Repositories {
						repos = append(repos, name)
					}
					sort.Strings(repos)
					m.repoChoices = repos
					m.state = StateViewRepos
					return m, nil

				case "Checkout Repo Branch":
					config, err := groveUtil.LoadConfig(m.path)
					if err != nil {
						m.err = err
						return m, nil
					}
					var repos []string
					for name := range config.Repositories {
						repos = append(repos, name)
					}
					sort.Strings(repos)
					m.repoChoices = repos
					m.repoCursor = 0
					m.state = StateRepoCheckoutSelection
					if len(repos) == 0 {
						m.err = fmt.Errorf("no repositories found")
					}
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

	case StateRegisterRepoName:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				name := m.textInput.Value()
				if name != "" {
					m.registerName = name
					m.state = StateRegisterRepoPath
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Enter repository path (relative to root)"

					// Initialize suggestions from root (m.path is repo root)
					m.suggestions = getSuggestions(m.path, "")
					m.suggestionCursor = -1
				}
			case tea.KeyEsc:
				m.state = StateIdle
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case StateRegisterRepoPath:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyTab: // Not tea.KeyTab? tea.KeyTab is correct but let's check standard tea
				if len(m.suggestions) > 0 {
					if m.suggestionCursor < 0 {
						m.suggestionCursor = 0
					}
					m.textInput.SetValue(m.suggestions[m.suggestionCursor]) // Use selected suggestion
					// Move cursor to end?
					m.textInput.CursorEnd()
					// Update suggestions based on new value.
					m.suggestions = getSuggestions(m.path, m.textInput.Value())
					m.suggestionCursor = -1
				}
				return m, nil
			case tea.KeyEnter:
				// If suggestions are active, select the highlighted suggestion
				if len(m.suggestions) > 0 && m.suggestionCursor >= 0 {
					selected := m.suggestions[m.suggestionCursor]

					if m.textInput.Value() != selected {
						m.textInput.SetValue(selected)
						m.textInput.CursorEnd()

						// Refresh suggestions for drill-down
						m.suggestions = getSuggestions(m.path, m.textInput.Value())
						m.suggestionCursor = -1
						return m, nil
					}
					// Fall through to submit
				}

				repoPath := m.textInput.Value()
				if repoPath != "" {
					// Call RegisterRepo
					newRepo := model.GGRepo{
						Name: m.registerName,
						Path: repoPath, // Should be relative path
					}
					// Only one repo
					if err := registerrepo.RegisterRepo([]model.GGRepo{newRepo}, m.path); err != nil {
						m.err = err
					} else {
						// Refresh context info
						currentBranch, _ := gitUtil.CurrentBranch(m.path)
						m.repoInfo = getTrunkContextInfo(m.path, currentBranch)
						m.state = StateIdle
					}
				}
			case tea.KeyEsc:
				m.state = StateIdle
				m.err = nil
			case tea.KeyUp:
				if len(m.suggestions) > 0 {
					m.suggestionCursor--
					if m.suggestionCursor < 0 {
						m.suggestionCursor = -1
					}
				}
				return m, nil
			case tea.KeyDown:
				if len(m.suggestions) > 0 {
					m.suggestionCursor++
					if m.suggestionCursor >= len(m.suggestions) {
						m.suggestionCursor = len(m.suggestions) - 1
					}
				}
				return m, nil
			default:
				// For any other key, update text input first
				m.textInput, cmd = m.textInput.Update(msg)
				// Then update suggestions
				m.suggestions = getSuggestions(m.path, m.textInput.Value())
				m.suggestionCursor = -1
				return m, cmd
			}
		}
		// If tab wasn't pressed
		return m, nil

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

	case StateViewRepos:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "enter", "q":
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
			}
		}

	case StateRepoCheckoutSelection:
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
					currentBranch, err := gitUtil.CurrentBranch(m.path)
					if err != nil {
						m.err = err
						return m, nil
					}
					targetBranch := fmt.Sprintf("gg/%s/%s", currentBranch, repoName)
					if err := gitUtil.Checkout(m.path, targetBranch); err != nil {
						m.err = fmt.Errorf("failed to checkout %s: %v", targetBranch, err)
					} else {
						m.isOrphan = true
						m.orphanRepoName = repoName
						m.trunkBranch = currentBranch
						m.repoInfo = fmt.Sprintf("Orphan Branch: %s (Trunk: %s)", repoName, currentBranch)
						m.choices = []string{"Prepare Merge", "Return to Trunk", "Quit"}
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
		lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("GitGrove"),
			infoStyle.Render("v1.0 ("+m.buildTime+")"),
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
