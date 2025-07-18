# Goatest 🚀

> **Integration testing for Go applications, as it should be!**

Goatest is a lightweight, powerful library that makes integration testing of Go applications effortless. Say goodbye to complex test setups, race conditions, and unreliable integration tests. With Goatest, you can run and test your Go applications in isolated environments with full control over their lifecycle.

## 🎯 Motivation

Integration testing should be simple, reliable, and fast. Too often, developers struggle with:

- **Complex test setups** - Manually managing application lifecycles, ports, and dependencies
- **Race conditions** - Unreliable tests that sometimes pass, sometimes fail
- **Environment management** - Difficulty in setting up different configurations for testing
- **Output capture** - Problems with capturing and asserting on application output
- **Process management** - Difficulty in properly starting and stopping test applications

Goatest solves all these problems by providing a clean, intuitive API that handles the complexity for you.

## ✨ Features

- 🚀 **Simple API** - Start and stop Go applications with a single function call
- 🔒 **Thread-safe** - Built-in synchronization prevents race conditions
- 🌍 **Environment management** - Easy configuration via environment variables and `.env` files
- 📝 **Output capture** - Capture, search, and assert on application output
- ⏱️ **Smart waiting** - Wait for specific conditions before proceeding with tests
- 🔄 **Process lifecycle** - Proper process management and cleanup
- 🎯 **Port management** - Automatic handling of different ports for parallel tests

## 📦 Installation

```bash
go get github.com/cappivara/goatest
```

## 🚀 Quick Start

### Basic Usage with TestMain

```go
package main

import (
    "net/http"
    "os"
    "strings"
    "testing"
    "time"

    "github.com/cappivara/goatest"
)

func TestMain(m *testing.M) {
    // Setup: Start the application before running tests
    p := &goatest.Process{
        File: "cmd/api/main.go",
        Env: map[string]string{
            "PORT": "8080",
            "ENV":  "test",
        },
        LogStream: os.Stdout,
        WaitingFor: func(output string) bool {
            return strings.Contains(output, "Server started on port 8080")
        },
    }

    if err := p.Run(); err != nil {
        panic("Failed to start test server: " + err.Error())
    }
    defer p.Stop() // Ensure cleanup happens

    // Run all tests
    os.Exit(m.Run())
}

func TestHealthEndpoint(t *testing.T) {
    resp, err := http.Get("http://localhost:8080/health")
    if err != nil {
        t.Fatalf("Health check failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}

func TestUserAPI(t *testing.T) {
    // Test user creation
    payload := `{"name": "John Doe", "email": "john@example.com"}`
    resp, err := http.Post("http://localhost:8080/users", "application/json", 
        strings.NewReader(payload))
    if err != nil {
        t.Fatalf("Create user failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        t.Errorf("Expected 201, got %d", resp.StatusCode)
    }
}
```

## 🌍 Environment Management Examples

### Using Environment Variables with TestMain

```go
package main

import (
    "os"
    "strings"
    "testing"
    "time"

    "github.com/cappivara/goatest"
)

func TestMain(m *testing.M) {
    // Start database service with environment variables
    p := &goatest.Process{
        File: "cmd/dbservice/main.go",
        Env: map[string]string{
            "DB_HOST":     "localhost",
            "DB_PORT":     "5432",
            "DB_NAME":     "test_db",
            "DB_USER":     "testuser",
            "DB_PASSWORD": "testpass",
            "API_PORT":    "9090",
            "LOG_LEVEL":   "debug",
        },
        LogStream: nil, // Capture output silently
        WaitingFor: func(output string) bool {
            return strings.Contains(output, "Database connected") &&
                   strings.Contains(output, "API server ready")
        },
    }

    if err := p.Run(); err != nil {
        panic("Failed to start database service: " + err.Error())
    }
    defer p.Stop() // Ensure cleanup happens

    // Wait a bit more for full initialization
    time.Sleep(2 * time.Second)

    os.Exit(m.Run())
}

func TestDatabaseOperations(t *testing.T) {
    // Your database tests here
    // The service is already running and connected to the database
}
```

### Using .env Files with TestMain

Create your test environment file:
```env
# test.env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=testdb
DB_USER=testuser
DB_PASSWORD=secret123
REDIS_URL=redis://localhost:6379
API_SECRET=test-secret-key
LOG_LEVEL=info
DEBUG=true
```

```go
package main

import (
    "os"
    "strings"
    "testing"

    "github.com/cappivara/goatest"
)

func TestMain(m *testing.M) {
    // Load configuration from .env file
    p := &goatest.Process{
        File:    "cmd/microservice/main.go",
        EnvFile: "test.env", // Load all variables from file
        LogStream: os.Stdout,
        WaitingFor: func(output string) bool {
            return strings.Contains(output, "All services initialized") &&
                   strings.Contains(output, "Health checks passed")
        },
    }

    if err := p.Run(); err != nil {
        panic("Failed to start microservice: " + err.Error())
    }
    defer p.Stop() // Ensure cleanup happens

    os.Exit(m.Run())
}

func TestMicroserviceEndpoints(t *testing.T) {
    // Test your microservice endpoints
    // All environment variables from test.env are loaded
}
```

### Environment Override with TestMain

```go
package main

import (
    "os"
    "strings"
    "testing"

    "github.com/cappivara/goatest"
)

func TestMain(m *testing.M) {
    // Load base config from .env, but override specific values
    p := &goatest.Process{
        File:    "cmd/server/main.go",
        EnvFile: "production.env", // Load production-like config
        Env: map[string]string{
            // Override specific values for testing
            "PORT":        "9999",           // Use different port for tests
            "ENV":         "test",           // Override environment
            "DB_NAME":     "test_database",  // Use test database
            "LOG_LEVEL":   "debug",          // More verbose logging for tests
            "RATE_LIMIT":  "1000",          // Higher rate limit for tests
        },
        LogStream: os.Stdout,
        WaitingFor: func(output string) bool {
            return strings.Contains(output, "Server listening on port 9999") &&
                   strings.Contains(output, "Environment: test")
        },
    }

    if err := p.Run(); err != nil {
        panic("Failed to start test server: " + err.Error())
    }
    defer p.Stop() // Ensure cleanup happens

    os.Exit(m.Run())
}

func TestServerConfiguration(t *testing.T) {
    // Note: In real scenarios, you'd need to access the process instance
    // or use HTTP requests to test the running server
    resp, err := http.Get("http://localhost:9999/config")
    if err != nil {
        t.Fatalf("Failed to get config: %v", err)
    }
    defer resp.Body.Close()
    // Test server configuration via HTTP endpoints
}
```

## 🔧 Advanced Testing Patterns with TestMain

### Multiple Services Setup

```go
package main

import (
    "os"
    "strings"
    "testing"
    "time"

    "github.com/cappivara/goatest"
)

func TestMain(m *testing.M) {
    // Start authentication service first
    authService := &goatest.Process{
        File: "cmd/auth/main.go",
        Env: map[string]string{
            "PORT":    "8081",
            "DB_NAME": "auth_test",
        },
        WaitingFor: func(output string) bool {
            return strings.Contains(output, "Auth service ready on port 8081")
        },
    }

    if err := authService.Run(); err != nil {
        panic("Failed to start auth service: " + err.Error())
    }
    defer authService.Stop() // Ensure cleanup happens

    // Start API server (depends on auth service)
    apiServer := &goatest.Process{
        File: "cmd/api/main.go",
        Env: map[string]string{
            "PORT":         "8080",
            "AUTH_SERVICE": "http://localhost:8081",
            "DB_NAME":      "api_test",
        },
        WaitingFor: func(output string) bool {
            return strings.Contains(output, "API server ready on port 8080")
        },
    }

    if err := apiServer.Run(); err != nil {
        panic("Failed to start API server: " + err.Error())
    }
    defer apiServer.Stop() // Ensure cleanup happens

    // Start background worker
    workerService := &goatest.Process{
        File: "cmd/worker/main.go",
        Env: map[string]string{
            "API_URL":    "http://localhost:8080",
            "WORKER_ID":  "test-worker-1",
        },
        WaitingFor: func(output string) bool {
            return strings.Contains(output, "Worker connected and ready")
        },
    }

    if err := workerService.Run(); err != nil {
        panic("Failed to start worker: " + err.Error())
    }
    defer workerService.Stop() // Ensure cleanup happens

    // Give everything a moment to fully initialize
    time.Sleep(2 * time.Second)

    // Run tests
    os.Exit(m.Run())
}

func TestFullSystemIntegration(t *testing.T) {
    // Test the entire system working together
    // All services are running and connected
}
```

### Dynamic Port Allocation

```go
package main

import (
    "fmt"
    "net"
    "os"
    "strings"
    "testing"

    "github.com/cappivara/goatest"
)

func TestMain(m *testing.M) {
    // Find an available port dynamically
    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        panic("Failed to find available port: " + err.Error())
    }
    testPort := fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
    listener.Close()

    p := &goatest.Process{
        File: "cmd/webapp/main.go",
        Env: map[string]string{
            "PORT": testPort,
            "ENV":  "test",
        },
        LogStream: os.Stdout,
        WaitingFor: func(output string) bool {
            return strings.Contains(output, fmt.Sprintf("Server ready on port %s", testPort))
        },
    }

    if err := p.Run(); err != nil {
        panic("Failed to start test app: " + err.Error())
    }
    defer p.Stop() // Ensure cleanup happens

    os.Exit(m.Run())
}

func TestWebApp(t *testing.T) {
    // Note: In real scenarios, you'd need to pass the port via environment
    // or discover it through service discovery
    resp, err := http.Get("http://localhost:8080/")
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }
    defer resp.Body.Close()
    // Test your web application
}
```

## 📚 Configuration Reference

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `File` | `string` | **Required.** Path to the Go file to run | `"cmd/server/main.go"` |
| `Env` | `map[string]string` | Environment variables to set | `map[string]string{"PORT": "8080"}` |
| `EnvFile` | `string` | Path to `.env` file to load | `".env"` or `"configs/test.env"` |
| `LogStream` | `io.Writer` | Where to write process output | `os.Stdout`, `os.Stderr`, `nil` |
| `WaitingFor` | `func(string) bool` | Function to check if process is ready | `func(s string) bool { return strings.Contains(s, "ready") }` |

### Environment Priority

When both `Env` and `EnvFile` are specified:
1. Variables from `EnvFile` are loaded first
2. Variables from `Env` override any duplicates from `EnvFile`
3. This allows you to have default configurations in files and override specific values for tests

### LogStream Options

- `os.Stdout` - Print output to console and capture it
- `os.Stderr` - Print output to stderr and capture it  
- `&strings.Builder{}` - Capture to a custom writer
- `nil` - Capture output silently (no console output)

## 🔧 Available Methods

### Process Control
- `Run() error` - Start the Go process
- `Stop()` - Stop the process and cleanup

### Output Management
- `GetOutput() string` - Get complete captured output
- `GetLines() []string` - Get output as individual lines
- `ContainsOutput(text string) bool` - Check if output contains text
- `WaitForOutput(text string, timeout time.Duration) bool` - Wait for specific output
- `ResetOutput()` - Clear captured output

## 🧪 Best Practices

### 1. Always Use TestMain for Integration Tests
```go
func TestMain(m *testing.M) {
    // Setup
    setupTestEnvironment()
    defer cleanupTestEnvironment() // Ensure cleanup happens
    
    // Run tests
    os.Exit(m.Run())
}
```

### 2. Use Environment Variables for Configuration
```go
// Good: Configurable and testable
testProcess := &goatest.Process{
    File: "cmd/app/main.go",
    Env: map[string]string{
        "PORT":    "8080",
        "DB_URL":  "postgres://localhost/testdb",
    },
}
```

### 3. Wait for Service Readiness
```go
// Good: Wait for actual readiness
WaitingFor: func(output string) bool {
    return strings.Contains(output, "Server ready") &&
           strings.Contains(output, "Database connected")
}
```

### 4. Use Unique Identifiers for Parallel Tests
```go
testID := fmt.Sprintf("test_%d", time.Now().UnixNano())
Env: map[string]string{
    "DB_NAME": "testdb_" + testID,
    "CACHE_PREFIX": testID,
}
```

## 🤝 Contributing

We welcome contributions! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

Built with ❤️ to make Go integration testing as smooth as it should be!