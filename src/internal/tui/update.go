package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	preparemerge "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/prepare-merge"
	registerrepo "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/register-repo"
	grovesync "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/sync"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
)

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
					m.textInput.SetValue("") // Start empty to show suggestions
					m.textInput.Placeholder = "Path to initialize (default: current directory)"

					// Initialize suggestions for CWD
					m.suggestions = getSuggestions(cwd, "")
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
			case tea.KeyTab:
				if len(m.suggestions) > 0 {
					if m.suggestionCursor < 0 {
						m.suggestionCursor = 0
					}
					m.textInput.SetValue(m.suggestions[m.suggestionCursor])
					m.textInput.CursorEnd()
					cwd, _ := os.Getwd()
					m.suggestions = getSuggestions(cwd, m.textInput.Value())
					m.suggestionCursor = -1
				}
				return m, nil

			case tea.KeyEnter:
				if len(m.suggestions) > 0 && m.suggestionCursor >= 0 {
					selected := m.suggestions[m.suggestionCursor]
					if m.textInput.Value() != selected {
						m.textInput.SetValue(selected)
						m.textInput.CursorEnd()
						cwd, _ := os.Getwd()
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

				if err != nil {
					// Check if error message confirms initialization
					if err.Error() == fmt.Sprintf("gitgrove is already initialized in %s", path) ||
						(len(err.Error()) > 30 && err.Error()[:31] == "gitgrove is already initialized") ||
						strings.Contains(err.Error(), "sticky context") {
						// Success
						m.path = path

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
							m.choices = []string{"Prepare Merge", "Sync from Trunk", "Return to Trunk", "Quit"}
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
				m.state = StateInit
				m.err = nil
				return m, nil

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
				m.textInput, cmd = m.textInput.Update(msg)
				cwd, _ := os.Getwd()
				m.suggestions = getSuggestions(cwd, m.textInput.Value())
				m.suggestionCursor = -1
				return m, cmd
			}
		}
		return m, nil

	case StateInputPath:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyTab:
				if len(m.suggestions) > 0 {
					if m.suggestionCursor < 0 {
						m.suggestionCursor = 0
					}
					m.textInput.SetValue(m.suggestions[m.suggestionCursor])
					m.textInput.CursorEnd()
					cwd, _ := os.Getwd()
					m.suggestions = getSuggestions(cwd, m.textInput.Value())
					m.suggestionCursor = -1
				}
				return m, nil

			case tea.KeyEnter:
				if len(m.suggestions) > 0 && m.suggestionCursor >= 0 {
					selected := m.suggestions[m.suggestionCursor]
					if m.textInput.Value() != selected {
						m.textInput.SetValue(selected)
						m.textInput.CursorEnd()
						cwd, _ := os.Getwd()
						m.suggestions = getSuggestions(cwd, m.textInput.Value())
						m.suggestionCursor = -1
						return m, nil
					}
				}

				path := m.textInput.Value()
				if path == "" {
					path, _ = os.Getwd()
				}

				if err := groveUtil.IsGroveInitialized(path); err != nil {
					m.err = err
					return m, nil
				}

				m.path = path
				m.state = StateInputAtomic
				return m, nil

			case tea.KeyEsc:
				m.state = StateInit
				m.err = nil
				return m, nil

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
				m.textInput, cmd = m.textInput.Update(msg)
				cwd, _ := os.Getwd()
				m.suggestions = getSuggestions(cwd, m.textInput.Value())
				m.suggestionCursor = -1
				return m, cmd
			}
		}
		return m, nil

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
					// Allow if isOrphan is true (covers sticky context)
					if m.isOrphan {
						// Pass m.orphanRepoName. If empty, PrepareMerge might fail or try sticky context again.
						// But m.orphanRepoName should be populated if isOrphan is true.
						if err := preparemerge.PrepareMerge(m.path, m.orphanRepoName); err != nil {
							m.err = err
						} else {
							m.repoInfo = "Success: Prepare-merge branch created"
							m.state = StateIdle
							// We might want to refresh model state here as checkouts happened?
							// The user is now on prepare-merge branch.
							// Sticky context should have been set by PrepareMerge.
							// So next TUI loop/restart will pick it up.
							// But for now, we just stay in IDLE.
						}
						return m, nil
					}

					m.err = fmt.Errorf("prepare merge only available in orphan branches")
					return m, nil

				case "Sync from Trunk":
					// Sync implementation:
					// 1. Check if orphan
					if !m.isOrphan {
						m.err = fmt.Errorf("sync only available in orphan branches")
						return m, nil
					}
					// 2. Call SyncOrphanFromTrunk
					if err := grovesync.SyncOrphanFromTrunk(m.path, "", m.trunkBranch, m.orphanRepoName); err != nil {
						m.err = err
					} else {
						m.repoInfo = "Success: Synced from trunk"
					}
					return m, nil

				case "Return to Orphan Branch":
					if m.orphanBranch == "" {
						m.err = fmt.Errorf("unknown orphan branch context")
						return m, nil
					}
					if err := gitUtil.Checkout(m.path, m.orphanBranch); err != nil {
						m.err = fmt.Errorf("failed to checkout orphan branch: %v", err)
					} else {
						// We don't clear context because we are still in the orphan context!
						// Just need to refresh view
						m.repoInfo = fmt.Sprintf("Returned to %s", m.orphanRepoName)
						// Refresh choices (will hide Return to Orphan since we are there)
						// We need to re-init choices logic?
						// Actually Model update happens in Update?
						// We need to valid choices again.
						// Similar to logic in NewModel... but we are updating m.
						// Let's duplicate basic choice logic here or trigger a state refresh?
						// For now manual update of choices:
						m.choices = []string{"Sync from Trunk", "Prepare Merge", "Return to Trunk", "Quit"}
						m.descriptions["Return to Orphan Branch"] = "" // Clear old if needed? Map persists.
						// Re-set default descriptions
						m.descriptions = map[string]string{
							"Sync from Trunk": "Merge latest changes from the trunk (for this component) into current branch.",
							"Prepare Merge":   "Prepares work for integration into the Trunk.",
							"Return to Trunk": "Switch back to main branch",
							"Quit":            "Exit the GitGrove application.",
						}
						m.state = StateIdle
					}
					return m, nil

				case "Return to Trunk":
					if m.trunkBranch == "" {
						m.err = fmt.Errorf("unknown trunk branch")
						return m, nil
					}
					if err := gitUtil.Checkout(m.path, m.trunkBranch); err != nil {
						m.err = fmt.Errorf("failed to checkout trunk: %v", err)
					} else {
						// Clear sticky context
						groveUtil.ClearAllContext(m.path)

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
						// Clean untracked files from previous context
						if err := gitUtil.Clean(m.path); err != nil {
							// Warning state? For now just log err or ignore?
							// Better to let user know?
							// Let's treat it as non-fatal but info.
							// For now, m.err = err?
							// But checkout succeeded.
							// Maybe append to repoInfo or a temp message?
							// Let's log to err to be safe.
							m.err = fmt.Errorf("checkout success, but failed to clean: %v", err)
						}

						// Set sticky context
						if err := groveUtil.SetContextRepo(m.path, repoName); err != nil {
							m.err = fmt.Errorf("checkout success, but failed to set context: %v", err)
						}
						// Set sticky trunk
						if err := groveUtil.SetContextTrunk(m.path, currentBranch); err != nil {
							// Log error but proceed?
						}
						// Set sticky orphan branch (the one we just checked out)
						if err := groveUtil.SetContextOrphan(m.path, targetBranch); err != nil {
							m.err = fmt.Errorf("checkout success, but failed to set orphan context: %v", err)
						}

						m.isOrphan = true
						m.orphanRepoName = repoName
						m.orphanBranch = targetBranch
						m.trunkBranch = currentBranch
						m.trunkBranch = currentBranch
						m.repoInfo = fmt.Sprintf("Orphan Branch: %s (Trunk: %s)", repoName, currentBranch)
						m.choices = []string{"Prepare Merge", "Sync from Trunk", "Return to Trunk", "Quit"}
						m.state = StateIdle
					}
				}
			}
		}
	}

	return m, nil
}
