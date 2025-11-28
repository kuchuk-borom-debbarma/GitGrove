package model

// Repo represents a repository registered in GitGroove.
type Repo struct {
	// Name is the unique identifier for the repo (e.g. "backend", "frontend").
	Name string `json:"name"`

	// Path is the relative path from the GitGroove root to the repo root.
	Path string `json:"path"`
}
