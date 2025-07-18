package goatest_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cappivara/goatest"
)

func TestProcessSingleRun(t *testing.T) {
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
}

func TestProcessWithEnvFile(t *testing.T) {
	builderOut := &strings.Builder{}
	r := goatest.Process{
		File:      "test/cmd/rest_api/main.go",
		EnvFile:   "test/data/.env.test",
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
}

func TestProcessRESTAPI(t *testing.T) {
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
	defer r.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:1010/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get http://localhost:1010/: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %d", resp.StatusCode)
	}
}

func TestProcessOutputAssertions(t *testing.T) {
	r := goatest.Process{
		File: "test/cmd/rest_api/main.go",
		Env: map[string]string{
			"PORT": "1011",
		},
		LogStream: os.Stdout,
	}

	if err := r.Run(); err != nil {
		t.Fatalf("failed to run: %v", err)
	}
	defer r.Stop()

	if !r.WaitForOutput("Server is running on port 1011", 5*time.Second) {
		t.Fatalf("expected output not found, got: %s", r.GetOutput())
	}

	if !r.ContainsOutput("Server is running on port 1011") {
		t.Fatalf("expected output to contain 'Server is running on port 1011', got: %s", r.GetOutput())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:1011/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get http://localhost:1011/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %d", resp.StatusCode)
	}
}

func TestProcessCaptureOnly(t *testing.T) {
	r := goatest.Process{
		File: "test/cmd/rest_api/main.go",
		Env: map[string]string{
			"PORT": "1012",
		},
		LogStream: nil,
	}

	if err := r.Run(); err != nil {
		t.Fatalf("failed to run: %v", err)
	}
	defer r.Stop()

	if !r.WaitForOutput("Server is running", 5*time.Second) {
		t.Fatalf("server did not start in time")
	}

	lines := r.GetLines()

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

	completeOutput := r.GetOutput()
	if !strings.Contains(completeOutput, "Server is running on port 1012") {
		t.Fatalf("complete output does not contain expected text: %s", completeOutput)
	}
}

func TestProcessWithWaitingFor(t *testing.T) {
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

	if !r.ContainsOutput("Server is running on port 1013") {
		t.Fatalf("expected output not found after WaitingFor condition: %s", r.GetOutput())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:1013/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get http://localhost:1013/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %d", resp.StatusCode)
	}
}

func TestProcessMultiplePatterns(t *testing.T) {
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

	if !r.WaitForOutput("Server is running", 5*time.Second) {
		t.Fatalf("server did not start in time, output: %s", r.GetOutput())
	}

	patterns := []string{
		"Server is running on port 1014",
		"1014",
	}

	for _, pattern := range patterns {
		if !r.ContainsOutput(pattern) {
			t.Fatalf("expected output to contain '%s', got: %s", pattern, r.GetOutput())
		}
	}
}

func TestProcessResetOutput(t *testing.T) {
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

	if !r.WaitForOutput("Server is running", 5*time.Second) {
		t.Fatalf("server did not start in time")
	}

	initialOutput := r.GetOutput()
	if initialOutput == "" {
		t.Fatal("expected some output before reset")
	}

	r.ResetOutput()

	if r.GetOutput() != "" {
		t.Fatalf("expected output to be cleared after reset, got: %s", r.GetOutput())
	}

	if len(r.GetLines()) != 0 {
		t.Fatalf("expected lines to be cleared after reset, got: %v", r.GetLines())
	}
}

func TestProcessEnvOverride(t *testing.T) {
	r := goatest.Process{
		File: "test/cmd/rest_api/main.go",
		Env: map[string]string{
			"PORT": "1016",
		},
		EnvFile:   "test/data/.env.test",
		LogStream: nil,
		WaitingFor: func(output string) bool {
			return strings.Contains(output, "Server is running on port 1016")
		},
	}

	if err := r.Run(); err != nil {
		t.Fatalf("failed to run: %v", err)
	}
	defer r.Stop()

	if !r.ContainsOutput("Server is running on port 1016") {
		t.Fatalf("expected Env to override EnvFile, got output: %s", r.GetOutput())
	}

	if r.ContainsOutput("Server is running on port 9999") {
		t.Fatalf("Env did not override EnvFile - still using EnvFile value, got output: %s", r.GetOutput())
	}
}
