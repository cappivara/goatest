package workdir

import (
	"fmt"
	"os"
	"path/filepath"
)

// MaxSearchIterations defines the maximum number of parent directories to search
// when looking for go.mod file to prevent infinite loops
const MaxSearchIterations = 30

// ProjectRoot recursively searches for go.mod file starting from the current directory
// and going up the directory tree. Returns the directory containing go.mod.
// Returns error if go.mod is not found after MaxSearchIterations or if no go.mod is found.
func ProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Resolve symlinks to get the real path
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	for i := 0; i < MaxSearchIterations; i++ {
		// Check if go.mod exists in current directory
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		// Move to parent directory
		parentDir := filepath.Dir(dir)

		// If we've reached the root and can't go further up
		if parentDir == dir {
			break
		}

		dir = parentDir
	}

	return "", fmt.Errorf("go.mod not found after %d iterations", MaxSearchIterations)
}
