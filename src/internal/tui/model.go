package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
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
	orphanBranch     string   // The original orphan branch name (e.g. gg/main/service-a)
	suggestions      []string // Autocompletion suggestions
	suggestionCursor int      // Selected suggestion index
	buildTime        string   // Build time of the binary
}

func InitialModel(buildTime string) Model {
	cwd, _ := os.Getwd()

	initialState := StateInit
	var repoInfo string
	var isOrphan bool
	var orphanRepoName, trunkBranch, orphanName, currentBranch string // Hoisted currentBranch

	// Check initialization status (returns error if initialized)
	errInit := groveUtil.IsGroveInitialized(cwd)

	if errInit != nil {
		initialState = StateIdle
		// Determine context: Trunk or Orphan?
		var err error
		currentBranch, err = gitUtil.CurrentBranch(cwd)
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
					// Inferred orphan branch name if we are on it
					orphanName = currentBranch
				} else {
					repoInfo = fmt.Sprintf("Orphan Branch: %s", currentBranch)
					orphanName = currentBranch
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

	// Try to overwrite with sticky context if available (more reliable for deep branches)
	if stickyOrphan, err := groveUtil.GetContextOrphan(cwd); err == nil && stickyOrphan != "" {
		orphanName = stickyOrphan
		// If we are deep, isOrphan might be false from prefix check, but sticky says we are in an orphan workflow.
		// So strict prefix check `if len(branch) > 3` above might be failing for `feat/foo`.
		// We should trust sticky context.
		if !isOrphan {
			isOrphan = true
			// We need to fetch other context too if not already set
			if stickyRepo, err := groveUtil.GetContextRepo(cwd); err == nil {
				orphanRepoName = stickyRepo
			}
			if stickyTrunk, err := groveUtil.GetContextTrunk(cwd); err == nil {
				trunkBranch = stickyTrunk
			}
			repoInfo = fmt.Sprintf("Feature Branch: %s (Root: %s)", currentBranch, orphanRepoName)
		}
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
			mainChoices = []string{"Sync from Trunk", "Prepare Merge", "Return to Trunk", "Quit"}
			descriptions["Sync from Trunk"] = "Merge latest changes from the trunk (for this component) into current branch."
			descriptions["Prepare Merge"] = "Prepare the current orphan branch for merging back into the trunk."
			descriptions["Return to Trunk"] = fmt.Sprintf("Checkout the trunk branch (%s) and leave the orphan state.", trunkBranch)

			// If we are deep in a feature branch (i.e., current branch != original orphan branch),
			// offer direct return to orphan.
			currentBranch, _ := gitUtil.CurrentBranch(cwd)
			// We have orphanName from earlier logic, but we need the full orphan branch name.
			// Earlier we set `m.orphanBranch`? No, we set top-level variable.
			// Let's ensure we use the local variable `orphanName` which we read from config.
			// Wait, in InitialModel we read `orphanName` from `GetContextOrphan`.

			// If we are relying on TUI detection logic:
			if isOrphan && currentBranch != strings.Join([]string{"gg", trunkBranch, orphanRepoName}, "/") {
				// Logic mismatch.
			}
			// Better: use the one we read from config.
			if orphanName != "" && currentBranch != orphanName {
				// Prepend or Append? "Return to Orphan Branch"
				// Let's put it before Return to Trunk
				mainChoices = []string{"Sync from Trunk", "Prepare Merge", "Return to Orphan Branch", "Return to Trunk", "Quit"}
				descriptions["Return to Orphan Branch"] = "Discard feature branch & return to component root"
			}
		} else {
			mainChoices = []string{"View Repos", "Register Repo", "Checkout Repo Branch", "Quit"}
			descriptions["View Repos"] = "View a list of all registered repositories in this workspace."
			descriptions["Register Repo"] = "Register a new repository (subdirectory) and create its orphan branch."
			descriptions["Checkout Repo Branch"] = "Switch context to a specific repository's orphan branch."
		}
	}
	descriptions["Quit"] = "Exit the GitGrove application."

	m := Model{
		state:            initialState,
		choices:          mainChoices,
		textInput:        ti,
		path:             cwd,
		repoInfo:         repoInfo,
		isOrphan:         isOrphan,
		orphanRepoName:   orphanRepoName,
		trunkBranch:      trunkBranch,
		orphanBranch:     orphanName,
		descriptions:     descriptions,
		suggestionCursor: -1,
		buildTime:        buildTime,
	}

	// Run an initial refresh to ensure all logic is consistent
	m.Refresh()
	return m
}

// Refresh updates the model state based on the current disk state
func (m *Model) Refresh() {
	// Only refresh if we are in a "viewing" state, not inputting text
	// Actually, context can change even while inputting, but we shouldn't change critical state that disrupts input.
	// For now, let's refresh repo info and branch status which are display-only mostly.

	cwd := m.path // Use current tracked path
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// Re-check init
	if err := groveUtil.IsGroveInitialized(cwd); err == nil {
		// Not initialized or error, maybe we lost init?
		// If we were initialized, this is a big change.
		// For safety, let's primarily check branch/context if we are already initialized.
		return
	}

	// We are initialized. Check branch context.
	currentBranch, err := gitUtil.CurrentBranch(cwd)
	if err != nil {
		return
	}

	var isOrphan bool
	var orphanRepoName, trunkBranch, orphanName string
	var repoInfo string

	// Check for Orphan Pattern
	if len(currentBranch) > 3 && currentBranch[:3] == "gg/" {
		isOrphan = true
		parts := strings.Split(currentBranch, "/")
		if len(parts) >= 3 {
			orphanRepoName = parts[len(parts)-1]
			trunkBranch = strings.Join(parts[1:len(parts)-1], "/")
			repoInfo = fmt.Sprintf("Orphan Branch: %s (Trunk: %s)", orphanRepoName, trunkBranch)
			orphanName = currentBranch
		} else {
			repoInfo = fmt.Sprintf("Orphan Branch: %s", currentBranch)
			orphanName = currentBranch
		}
	} else {
		repoInfo = getTrunkContextInfo(cwd, currentBranch)
	}

	// Sticky context check
	if stickyOrphan, err := groveUtil.GetContextOrphan(cwd); err == nil && stickyOrphan != "" {
		orphanName = stickyOrphan
		if !isOrphan {
			isOrphan = true
			if stickyRepo, err := groveUtil.GetContextRepo(cwd); err == nil {
				orphanRepoName = stickyRepo
			}
			if stickyTrunk, err := groveUtil.GetContextTrunk(cwd); err == nil {
				trunkBranch = stickyTrunk
			}
			repoInfo = fmt.Sprintf("Feature Branch: %s (Root: %s)", currentBranch, orphanRepoName)
		}
	}

	// Update model
	m.isOrphan = isOrphan
	m.repoInfo = repoInfo
	if isOrphan {
		m.orphanRepoName = orphanRepoName
		m.trunkBranch = trunkBranch
		m.orphanBranch = orphanName
	}

	// Update choices based on new state ONLY if we are in StateIdle to avoid disrupting menus
	if m.state == StateIdle {
		if isOrphan {
			// Update orphan choices
			// Preserve correctness
			if m.orphanBranch != "" && currentBranch != m.orphanBranch {
				m.choices = []string{"Sync from Trunk", "Prepare Merge", "Return to Orphan Branch", "Return to Trunk", "Quit"}
				m.descriptions["Return to Orphan Branch"] = "Discard feature branch & return to component root"
			} else {
				m.choices = []string{"Sync from Trunk", "Prepare Merge", "Return to Trunk", "Quit"}
				delete(m.descriptions, "Return to Orphan Branch")
			}
		} else {
			// Trunk choices
			m.choices = []string{"View Repos", "Register Repo", "Checkout Repo Branch", "Quit"}
		}
	}
}
