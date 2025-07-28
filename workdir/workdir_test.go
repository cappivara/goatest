package workdir

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectRoot_FindsGoModInCurrentDir(t *testing.T) {
	// Create temporary directory with go.mod
	tempDir, err := os.MkdirTemp("", "goatest-workdir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create go.mod in temp directory
	goModPath := filepath.Join(tempDir, "go.mod")
	if writeErr := os.WriteFile(goModPath, []byte("module test\n"), 0644); writeErr != nil {
		t.Fatalf("failed to create go.mod: %v", writeErr)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	defer func() {
		if chDirErr := os.Chdir(originalDir); chDirErr != nil {
			t.Errorf("failed to restore original directory: %v", chDirErr)
		}
	}()

	if chDirErr := os.Chdir(tempDir); chDirErr != nil {
		t.Fatalf("failed to change dir: %v", chDirErr)
	}

	// Test ProjectRoot
	root, err := ProjectRoot()
	if err != nil {
		t.Fatalf("ProjectRoot() failed: %v", err)
	}

	// Resolve symlinks for comparison
	expectedPath, evalErr := filepath.EvalSymlinks(tempDir)
	if evalErr != nil {
		t.Fatalf("failed to resolve symlinks: %v", evalErr)
	}

	if root != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, root)
	}
}

func TestProjectRoot_FindsGoModInParentDir(t *testing.T) {
	// Create temporary directory structure: tempDir/go.mod, tempDir/subdir/
	tempDir, err := os.MkdirTemp("", "goatest-workdir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create go.mod in root temp directory
	goModPath := filepath.Join(tempDir, "go.mod")
	if writeErr := os.WriteFile(goModPath, []byte("module test\n"), 0644); writeErr != nil {
		t.Fatalf("failed to create go.mod: %v", writeErr)
	}

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if mkDirErr := os.Mkdir(subDir, 0755); mkDirErr != nil {
		t.Fatalf("failed to create subdir: %v", mkDirErr)
	}

	// Change to subdirectory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	defer func() {
		if chDirErr := os.Chdir(originalDir); chDirErr != nil {
			t.Errorf("failed to restore original directory: %v", chDirErr)
		}
	}()

	if chDirErr := os.Chdir(subDir); chDirErr != nil {
		t.Fatalf("failed to change dir: %v", chDirErr)
	}

	// Test ProjectRoot
	root, err := ProjectRoot()
	if err != nil {
		t.Fatalf("ProjectRoot() failed: %v", err)
	}

	// Resolve symlinks for comparison
	expectedPath, evalErr := filepath.EvalSymlinks(tempDir)
	if evalErr != nil {
		t.Fatalf("failed to resolve symlinks: %v", evalErr)
	}

	if root != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, root)
	}
}

func TestProjectRoot_ReturnsErrorAfter30Iterations(t *testing.T) {
	// Create deep directory structure without go.mod
	tempDir, err := os.MkdirTemp("", "goatest-workdir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested directories (more than MaxSearchIterations limit)
	currentDir := tempDir
	for i := 0; i < MaxSearchIterations+5; i++ {
		currentDir = filepath.Join(currentDir, "level")
		if mkDirErr := os.MkdirAll(currentDir, 0755); mkDirErr != nil {
			t.Fatalf("failed to create nested dir: %v", mkDirErr)
		}
	}

	// Change to deepest directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	defer func() {
		if chDirErr := os.Chdir(originalDir); chDirErr != nil {
			t.Errorf("failed to restore original directory: %v", chDirErr)
		}
	}()

	if chDirErr := os.Chdir(currentDir); chDirErr != nil {
		t.Fatalf("failed to change dir: %v", chDirErr)
	}

	// Test ProjectRoot should return error
	_, projectErr := ProjectRoot()
	if projectErr == nil {
		t.Errorf("expected error after %d iterations, got nil", MaxSearchIterations)
	}

	expectedMsg := fmt.Sprintf("go.mod not found after %d iterations", MaxSearchIterations)
	if projectErr.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, projectErr.Error())
	}
}

func TestProjectRoot_ReturnsErrorWhenNoGoModFound(t *testing.T) {
	// Create temporary directory without go.mod
	tempDir, err := os.MkdirTemp("", "goatest-workdir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	defer func() {
		if chDirErr := os.Chdir(originalDir); chDirErr != nil {
			t.Errorf("failed to restore original directory: %v", chDirErr)
		}
	}()

	if chDirErr := os.Chdir(tempDir); chDirErr != nil {
		t.Fatalf("failed to change dir: %v", chDirErr)
	}

	// Test ProjectRoot should return error
	_, projectErr := ProjectRoot()
	if projectErr == nil {
		t.Error("expected error when no go.mod found, got nil")
	}
}
