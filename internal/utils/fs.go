package utils

import (
	"path/filepath"
	"strings"
)

// IsIn checks if a path is in a directory
// It returns true if the path is the directory itself
func IsIn(path string, dir string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}

    if absPath == absDir {
        return true
    }

	relPath, err := filepath.Rel(absDir, absPath)
	if err != nil {
		return false
	}

	return len(relPath) != 0 &&
		!strings.HasPrefix(relPath, "..")
}

