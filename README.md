# Goatest

A Go library for integration testing without the boilerplate. No more complex initialization logic or duplicated setup code in your tests. Simply run and test your Go applications with full control over environment, output, and lifecycle.

## Installation

```bash
go get github.com/cappivara/goatest
```

## Usage

```go
func TestMain(m *testing.M) {
    // Start your application
    p := &goatest.Process{
        File:    "cmd/server/main.go",
        EnvFile: ".env.test",
        Env: map[string]string{
            "PORT": "8080",        // Env overrides EnvFile
            "ENV":  "test",
        },
        LogStream: os.Stdout,      // or nil to capture silently
        WaitingFor: func(output string) bool {
            return strings.Contains(output, "Server ready")
        },
    }

    if err := p.Run(); err != nil {
        panic("Failed to start server: " + err.Error())
    }
    
    code := m.Run()
    
    p.Stop()
    os.Exit(code)
}

func TestAPI(t *testing.T) {
    // Your server is running and ready for testing
    resp, err := http.Get("http://localhost:8080/health")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}
```

## Features

- **Process Management** - Start and stop Go applications in tests
- **Environment Control** - Set environment variables via `Env` map or `.env` files
- **Output Capture** - Capture and assert on application output
- **Ready Detection** - Wait for specific output before proceeding
- **Thread-Safe** - Safe for concurrent test execution

## API

### Process Configuration

| Field | Type | Description |
|-------|------|-------------|
| `File` | `string` | Path to the Go file to run |
| `Env` | `map[string]string` | Environment variables (overrides EnvFile) |
| `EnvFile` | `string` | Path to `.env` file |
| `LogStream` | `io.Writer` | Where to write output (nil = capture only) |
| `WaitingFor` | `func(string) bool` | Ready condition checker |

### Methods

- `Run() error` - Start the process
- `Stop()` - Stop the process
- `GetOutput() string` - Get captured output
- `ContainsOutput(string) bool` - Check if output contains text
- `WaitForOutput(string, time.Duration) bool` - Wait for specific output

## License

MIT
