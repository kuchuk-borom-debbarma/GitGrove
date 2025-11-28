package core

import (
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
)

// Init initializes GitGroove on the current Git repository.
// It delegates to the internal grove package.
func Init(absolutePath string) error {
	return grove.Init(absolutePath)
}

// Register records one or more repos (name → path) in the GitGroove metadata.
func Register(rootAbsPath string, repos map[string]string) error {
	return grove.Register(rootAbsPath, repos)
}

// Link defines one or more repo hierarchy relationships (childName → parentName).
func Link(rootAbsPath string, relationships map[string]string) error {
	return grove.Link(rootAbsPath, relationships)
}
