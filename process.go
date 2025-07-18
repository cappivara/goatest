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
func (p *Process) GetOutput() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.safeWriter != nil {
		return p.safeWriter.getOutput()
	}
	return ""
}

// GetLines returns all captured output lines
func (p *Process) GetLines() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.safeWriter != nil {
		return p.safeWriter.getLines()
	}
	return nil
}

// ContainsOutput checks if the output contains the given string
func (p *Process) ContainsOutput(text string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.safeWriter != nil {
		return p.safeWriter.containsOutput(text)
	}
	return false
}

// WaitForOutput waits for specific output to appear (with timeout)
func (p *Process) WaitForOutput(text string, timeout time.Duration) bool {
	p.mu.Lock()
	safeWriter := p.safeWriter
	p.mu.Unlock()

	if safeWriter != nil {
		return safeWriter.waitForOutput(text, timeout)
	}
	return false
}

// ResetOutput clears the captured output
func (p *Process) ResetOutput() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.safeWriter != nil {
		p.safeWriter.reset()
	}
}

// Run starts the Go process in the background
func (p *Process) Run() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return nil
	}

	if p.File == "" {
		return fmt.Errorf("no file specified")
	}

	// Initialize Env map if nil
	if p.Env == nil {
		p.Env = make(map[string]string)
	}

	// Load environment variables from EnvFile if specified
	if p.EnvFile != "" {
		if err := p.loadEnvFile(); err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}
	}

	// Wrap LogStream with thread-safe wrapper
	p.safeWriter = newThreadSafeWriter(p.LogStream)

	// Build the go run command
	p.cmd = exec.Command("go", "run", p.File)

	// Set environment variables
	p.cmd.Env = os.Environ()
	for k, v := range p.Env {
		p.cmd.Env = append(p.cmd.Env, k+"="+v)
	}

	// Set working directory to the current directory (runner)
	if wd, err := os.Getwd(); err == nil {
		p.cmd.Dir = wd
	}

	// Set process group to enable killing the entire process tree
	p.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Create pipes for stdout and stderr
	stdoutPipe, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := p.cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the process
	if err := p.cmd.Start(); err != nil {
		return err
	}

	p.running = true

	// Channel to signal when the waiting condition is met
	readyChan := make(chan bool, 1)

	// Function to handle output lines
	handleLine := func(line string) {
		// Write to thread-safe writer (which handles LogStream internally)
		if p.safeWriter != nil {
			_, _ = p.safeWriter.Write([]byte(line))
		}

		// Check waiting condition if set using the thread-safe writer's output
		if p.WaitingFor != nil && p.safeWriter != nil && p.WaitingFor(p.safeWriter.getOutput()) {
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
	if p.WaitingFor != nil {
		// Log that we're waiting for readiness
		if p.safeWriter != nil {
			_, _ = p.safeWriter.Write([]byte("[runner] Waiting for the readiness.\n"))
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
		_ = p.cmd.Wait()
	}()

	return nil
}

// Stop terminates the background process
func (p *Process) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running || p.cmd == nil {
		return
	}

	// Kill the entire process group to ensure child processes are also terminated
	if p.cmd.Process != nil {
		// Get the process group ID
		pgid, err := syscall.Getpgid(p.cmd.Process.Pid)
		if err == nil {
			// Kill the entire process group (negative PID means process group)
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			// Fallback: kill just the main process
			_ = p.cmd.Process.Kill()
		}
	}

	p.running = false
}

// loadEnvFile loads environment variables from a .env file
func (p *Process) loadEnvFile() error {
	file, err := os.Open(p.EnvFile)
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
		if _, exists := p.Env[key]; !exists {
			p.Env[key] = value
		}
	}

	return scanner.Err()
}
