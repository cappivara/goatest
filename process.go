package goatest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Process manages the lifecycle of a background Go process
type Process struct {
	File       string
	Env        map[string]string
	LogStream  io.Writer
	WaitingFor func(string) bool

	// private fields for internal state
	cmd     *exec.Cmd
	mu      sync.Mutex
	running bool
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

	// Setup output handling
	var outputBuffer bytes.Buffer
	var bufferMu sync.Mutex

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
		bufferMu.Lock()
		outputBuffer.WriteString(line)
		currentOutput := outputBuffer.String()
		bufferMu.Unlock()

		// Write to log stream if configured
		if r.LogStream != nil {
			_, _ = r.LogStream.Write([]byte(line))
		}

		// Check waiting condition if set
		if r.WaitingFor != nil && r.WaitingFor(currentOutput) {
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
		if r.LogStream != nil {
			_, _ = r.LogStream.Write([]byte("[runner] Waiting for the readiness.\n"))
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
