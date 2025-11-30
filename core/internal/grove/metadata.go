package grove

import (
	"fmt"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

// Metadata represents the state of GitGrove metadata at a specific commit.
type Metadata struct {
	Commit string                // The commit hash this metadata was loaded from
	Repos  map[string]model.Repo // All registered repositories
}

// MetadataService provides a centralized way to load and update GitGrove metadata.
type MetadataService struct {
	rootPath string
}

// NewMetadataService creates a new metadata service for the given repository root.
func NewMetadataService(rootPath string) *MetadataService {
	return &MetadataService{rootPath: rootPath}
}

// Load reads the current metadata from the internal branch.
func (m *MetadataService) Load() (*Metadata, error) {
	// Resolve internal branch
	tip, err := gitUtil.ResolveRef(m.rootPath, InternalBranchRef)
	if err != nil {
		return nil, fmt.Errorf("not initialized or internal branch missing: %w", err)
	}

	// Load repositories
	repos, err := loadExistingRepos(m.rootPath, tip)
	if err != nil {
		return nil, fmt.Errorf("failed to load repositories: %w", err)
	}

	return &Metadata{
		Commit: tip,
		Repos:  repos,
	}, nil
}

// Update applies changes to metadata atomically.
// The updateFn receives the current metadata and should modify it as needed.
// Returns the new commit hash.
func (m *MetadataService) Update(updateFn func(*Metadata) error, message string) (string, error) {
	// Load current metadata
	meta, err := m.Load()
	if err != nil {
		return "", err
	}

	// Apply updates
	if err := updateFn(meta); err != nil {
		return "", fmt.Errorf("update function failed: %w", err)
	}

	// TODO: Implement commit logic
	// This would require converting the Metadata back to file updates
	// For now, this is a placeholder for future enhancement
	return meta.Commit, fmt.Errorf("Update not yet implemented - use direct operations")
}
