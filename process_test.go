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
	t.Run("should run a single process", func(t *testing.T) {
		r := goatest.Process{
			File: "test/cmd/single_process/main.go",
			Env: map[string]string{
				"PORT": "1010",
			},
			LogStream: os.Stdout,
		}

		if err := r.Run(); err != nil {
			t.Fatalf("failed to run: %v", err)
		}

		time.Sleep(10 * time.Second)

		r.Stop()
	})

	t.Run("rest api should be available", func(t *testing.T) {
		r := goatest.Process{
			File: "cmd/rest_api/main.go",
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
}
