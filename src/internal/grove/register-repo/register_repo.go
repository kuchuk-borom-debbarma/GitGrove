package registerrepo

// RegisterRepo registers a folder as a "repo" within the GitGrove monorepo.
//
// Concept: The Split
// When a folder is registered as a "repo," GG creates a parallel history for it.
//
// Workflow:
//  1. Orphan Creation: GG scans the history of the target folder.
//  2. Path Translation: It uses git subtree split to create a new orphan branch (e.g., gg/serviceA).
//  3. Root Projection: Inside this new branch, the files are moved to the Root Directory.
//     - Trunk View: ./backend/services/serviceA/main.go
//     - Orphan View: ./main.go
//
// Limitation: Nested Repositories
// Nested directories cannot be registered as repositories at this time.
// Rules for this are yet to be clearly defined.
func RegisterRepo(repoName, path string) error {
	return nil
}
