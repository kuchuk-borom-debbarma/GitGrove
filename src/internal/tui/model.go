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
