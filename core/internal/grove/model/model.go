package model

// Repo represents a repository registered in GitGroove.
//
// This struct is serialized to JSON and stored in .gg/repos/<name>/path (currently just path string,
// but this model supports future extensibility).
//
// Usage:
//   - Loaded from .gg/repos during Init/Register/List operations.
//   - Used to map logical names to physical paths.
type Repo struct {
	// Name is the unique identifier for the repo (e.g. "backend", "frontend").
	// It must be unique within the GitGroove project.
	Name string `json:"name"`

	// Path is the relative path from the GitGroove root to the repo root.
	// It uses forward slashes ("/") as separators.
	Path string `json:"path"`

	// Parent is the name of the parent repository in the hierarchy.
	// Empty if it's a root repository.
	Parent string `json:"parent,omitempty"`
}

const DefaultRepoBranch = "main"
