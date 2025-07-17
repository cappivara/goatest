package goatest_test

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cappivara/goatest"
)

func TestProcess(t *testing.T) {
	t.Run("run a single process", func(t *testing.T) {
		out := &strings.Builder{}

		r := goatest.Process{
			File: "test/cmd/single_process/main.go",
			Env: map[string]string{
				"PORT": "1010",
			},
			LogStream: out,
			WaitingFor: func(output string) bool {
				return strings.Contains(output, "Hello, World! 1010")
			},
		}

		if err := r.Run(); err != nil {
			t.Fatalf("failed to run: %v", err)
		}
		defer r.Stop()

		if !strings.Contains(out.String(), "Hello, World! 1010") {
			t.Fatalf("expected output 'Hello, World! 1010', got: %s", out.String())
		}
	})

	t.Run("run with envfile", func(t *testing.T) {
		builderOut := &strings.Builder{}
		r := goatest.Process{
			File:      "test/cmd/rest_api/main.go",
			EnvFile:   "test/data/.env",
			LogStream: builderOut,
			WaitingFor: func(s string) bool {
				return strings.Contains(s, "Server is running on port 9999")
			},
		}
		defer r.Stop()

		if err := r.Run(); err != nil {
			t.Errorf("error on run the process: %v", err.Error())
		}

		if got, want := builderOut.String(), "Server is running on port 9999"; !strings.Contains(got, want) {
			t.Fatalf("got message \"%s\", want: \"%s\"", got, want)
		}
	})

	t.Run("rest api should be available", func(t *testing.T) {
		r := goatest.Process{
			File: "test/cmd/rest_api/main.go",
			Env: map[string]string{
				"PORT": "1010",
			},
			LogStream: os.Stdout,
			WaitingFor: func(output string) bool {
				return strings.Contains(output, "Server is running on port 1010")
			},
		}

		if err := r.Run(); err != nil {
			t.Fatalf("failed to run: %v", err)
		}

		resp, err := http.Get("http://localhost:1010/")
		if err != nil {
			t.Fatalf("failed to get http://localhost:1010/: %v", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status OK, got %d", resp.StatusCode)
		}

		r.Stop()
	})

	// Strategy 1: Using Process with LogStream for safe output assertions
	t.Run("should automatically wrap LogStream for safe output assertions", func(t *testing.T) {
		r := goatest.Process{
			File: "test/cmd/rest_api/main.go",
			Env: map[string]string{
				"PORT": "1011",
			},
			LogStream: os.Stdout, // User passes any io.Writer
		}

		if err := r.Run(); err != nil {
			t.Fatalf("failed to run: %v", err)
		}
		defer r.Stop()

		// Wait for the expected output with timeout - no race condition!
		if !r.WaitForOutput("Server is running on port 1011", 5*time.Second) {
			t.Fatalf("expected output not found, got: %s", r.GetOutput())
		}

		// Assert the output contains expected text
		if !r.ContainsOutput("Server is running on port 1011") {
			t.Fatalf("expected output to contain 'Server is running on port 1011', got: %s", r.GetOutput())
		}

		// Test HTTP endpoint is working
		resp, err := http.Get("http://localhost:1011/")
		if err != nil {
			t.Fatalf("failed to get http://localhost:1011/: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status OK, got %d", resp.StatusCode)
		}
	})

	// Strategy 2: Using Process without LogStream (capture only)
	t.Run("should capture output without printing to stdout", func(t *testing.T) {
		r := goatest.Process{
			File: "test/cmd/rest_api/main.go",
			Env: map[string]string{
				"PORT": "1012",
			},
			LogStream: nil, // No output to stdout
		}

		if err := r.Run(); err != nil {
			t.Fatalf("failed to run: %v", err)
		}
		defer r.Stop()

		// Wait for output with timeout
		if !r.WaitForOutput("Server is running", 5*time.Second) {
			t.Fatalf("server did not start in time")
		}

		// Get all output lines for detailed assertions
		lines := r.GetLines()

		// Check if we have the expected line
		found := false
		for _, line := range lines {
			if strings.Contains(line, "Server is running on port 1012") {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("expected line not found in output: %v", lines)
		}

		// Verify the complete output
		completeOutput := r.GetOutput()
		if !strings.Contains(completeOutput, "Server is running on port 1012") {
			t.Fatalf("complete output does not contain expected text: %s", completeOutput)
		}
	})

	// Strategy 3: Combining Process with WaitingFor
	t.Run("should combine automatic wrapping with WaitingFor", func(t *testing.T) {
		r := goatest.Process{
			File: "test/cmd/rest_api/main.go",
			Env: map[string]string{
				"PORT": "1013",
			},
			LogStream: os.Stdout,
			WaitingFor: func(output string) bool {
				return strings.Contains(output, "Server is running on port 1013")
			},
		}

		if err := r.Run(); err != nil {
			t.Fatalf("failed to run: %v", err)
		}
		defer r.Stop()

		// Since WaitingFor returned true, we know the output is captured
		// Additional assertions after WaitingFor condition is met
		if !r.ContainsOutput("Server is running on port 1013") {
			t.Fatalf("expected output not found after WaitingFor condition: %s", r.GetOutput())
		}

		// Test that the server is actually running
		resp, err := http.Get("http://localhost:1013/")
		if err != nil {
			t.Fatalf("failed to get http://localhost:1013/: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status OK, got %d", resp.StatusCode)
		}
	})

	// Strategy 4: Testing multiple output patterns with automatic wrapping
	t.Run("should handle multiple output patterns safely", func(t *testing.T) {
		r := goatest.Process{
			File: "test/cmd/rest_api/main.go",
			Env: map[string]string{
				"PORT": "1014",
			},
			LogStream: os.Stdout,
		}

		if err := r.Run(); err != nil {
			t.Fatalf("failed to run: %v", err)
		}
		defer r.Stop()

		// Wait for server to start
		if !r.WaitForOutput("Server is running", 5*time.Second) {
			t.Fatalf("server did not start in time, output: %s", r.GetOutput())
		}

		// Test multiple patterns
		patterns := []string{
			"Server is running on port 1014",
			"1014",
		}

		for _, pattern := range patterns {
			if !r.ContainsOutput(pattern) {
				t.Fatalf("expected output to contain '%s', got: %s", pattern, r.GetOutput())
			}
		}
	})

	// Strategy 5: Using Reset functionality
	t.Run("should be able to reset captured output", func(t *testing.T) {
		r := goatest.Process{
			File: "test/cmd/rest_api/main.go",
			Env: map[string]string{
				"PORT": "1015",
			},
			LogStream: nil,
		}

		if err := r.Run(); err != nil {
			t.Fatalf("failed to run: %v", err)
		}
		defer r.Stop()

		// Wait for initial output
		if !r.WaitForOutput("Server is running", 5*time.Second) {
			t.Fatalf("server did not start in time")
		}

		// Verify we have output
		initialOutput := r.GetOutput()
		if len(initialOutput) == 0 {
			t.Fatal("expected some output before reset")
		}

		// Reset the output
		r.ResetOutput()

		// Verify output is cleared
		if len(r.GetOutput()) != 0 {
			t.Fatalf("expected output to be cleared after reset, got: %s", r.GetOutput())
		}

		if len(r.GetLines()) != 0 {
			t.Fatalf("expected lines to be cleared after reset, got: %v", r.GetLines())
		}
	})
}
