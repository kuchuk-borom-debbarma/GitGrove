package core

import (
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
)

// Init initializes GitGroove on the current Git repository.
// It delegates to the internal grove package.
func Init(absolutePath string) error {
	return grove.Init(absolutePath)
}
