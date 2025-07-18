package goatest

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Process manages the lifecycle of a background Go process
type Process struct {
	File       string
	Env        map[string]string
	EnvFile    string
	LogStream  io.Writer
	WaitingFor func(string) bool

	// private fields for internal state
	cmd        *exec.Cmd
	mu         sync.Mutex
	running    bool
	safeWriter *threadSafeWriter
}

// GetOutput returns the complete captured output
func (r *Process) GetOutput() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.safeWriter != nil {
		return r.safeWriter.getOutput()
	}
	return ""
}

// GetLines returns all captured output lines
func (r *Process) GetLines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.safeWriter != nil {
		return r.safeWriter.getLines()
	}
	return nil
}

// ContainsOutput checks if the output contains the given string
func (r *Process) ContainsOutput(text string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.safeWriter != nil {
		return r.safeWriter.containsOutput(text)
	}
	return false
}

// WaitForOutput waits for specific output to appear (with timeout)
func (r *Process) WaitForOutput(text string, timeout time.Duration) bool {
	r.mu.Lock()
	safeWriter := r.safeWriter
	r.mu.Unlock()

	if safeWriter != nil {
		return safeWriter.waitForOutput(text, timeout)
	}
	return false
}

// ResetOutput clears the captured output
func (r *Process) ResetOutput() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.safeWriter != nil {
		r.safeWriter.reset()
	}
}

// Run starts the Go process in the background
func (r *Process) Run() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return nil
	}

	if r.File == "" {
		return fmt.Errorf("no file specified")
	}

	// Initialize Env map if nil
	if r.Env == nil {
		r.Env = make(map[string]string)
	}

	// Load environment variables from EnvFile if specified
	if r.EnvFile != "" {
		if err := r.loadEnvFile(); err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}
	}

	// Wrap LogStream with thread-safe wrapper
	r.safeWriter = newThreadSafeWriter(r.LogStream)

	// Build the go run command
	r.cmd = exec.Command("go", "run", r.File)

	// Set environment variables
	r.cmd.Env = os.Environ()
	for k, v := range r.Env {
		r.cmd.Env = append(r.cmd.Env, k+"="+v)
	}

	// Set working directory to the current directory (runner)
	if wd, err := os.Getwd(); err == nil {
		r.cmd.Dir = wd
	}

	// Set process group to enable killing the entire process tree
	r.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Create pipes for stdout and stderr
	stdoutPipe, err := r.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := r.cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the process
	if err := r.cmd.Start(); err != nil {
		return err
	}

	r.running = true

	// Channel to signal when the waiting condition is met
	readyChan := make(chan bool, 1)

	// Function to handle output lines
	handleLine := func(line string) {
		// Write to thread-safe writer (which handles LogStream internally)
		if r.safeWriter != nil {
			_, _ = r.safeWriter.Write([]byte(line))
		}

		// Check waiting condition if set using the thread-safe writer's output
		if r.WaitingFor != nil && r.safeWriter != nil && r.WaitingFor(r.safeWriter.getOutput()) {
			select {
			case readyChan <- true:
			default:
			}
		}
	}

	// Handle stdout
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			handleLine(scanner.Text() + "\n")
		}
	}()

	// Handle stderr
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			handleLine(scanner.Text() + "\n")
		}
	}()

	// Wait for the condition if specified
	if r.WaitingFor != nil {
		// Log that we're waiting for readiness
		if r.safeWriter != nil {
			_, _ = r.safeWriter.Write([]byte("[runner] Waiting for the readiness.\n"))
		}

		select {
		case <-readyChan:
			// Condition met, continue
		case <-time.After(30 * time.Second):
			// Timeout after 30 seconds
		}
	}

	// Continue running the process in background
	go func() {
		_ = r.cmd.Wait()
	}()

	return nil
}

// Stop terminates the background process
func (r *Process) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running || r.cmd == nil {
		return
	}

	// Kill the entire process group to ensure child processes are also terminated
	if r.cmd.Process != nil {
		// Get the process group ID
		pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
		if err == nil {
			// Kill the entire process group (negative PID means process group)
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			// Fallback: kill just the main process
			_ = r.cmd.Process.Kill()
		}
	}

	r.running = false
}

// loadEnvFile loads environment variables from a .env file
func (r *Process) loadEnvFile() error {
	file, err := os.Open(r.EnvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Split by first = to handle values with = in them
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}
		
		// Only set if key doesn't already exist (Env should override EnvFile)
		if _, exists := r.Env[key]; !exists {
			r.Env[key] = value
		}
	}
	
	return scanner.Err()
}
