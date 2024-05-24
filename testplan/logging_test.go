package testplan

import (
	"os"
	"testing"
)

func TestLogging(t *testing.T) {
	if os.Getenv("GITHUB_ENV") == "" {
		// running locally
		os.Setenv("GITHUB_ENV", "../env.tmp")
	}

	levels := [7]string{"PANIC", "FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}
	for _, l := range levels {
		t.Run("Test loglevel", func(t *testing.T) {
			t.Setenv("INPUT_FILES", "../example/defaults.yaml")
			t.Setenv("INPUT_LOGLEVEL", l)
			plan, err := New()
			if err != nil {
				t.Logf("Got error %v", err)
				t.Errorf("Error when loading yaml files")
			}
			if plan.LogLevel != l {
				t.Errorf("Expected '%v' as LogLevel, got %v", l, plan.LogLevel)
			}
		})
	}

	t.Run("Test wrong loglevel", func(t *testing.T) {
		t.Setenv("INPUT_FILES", "../example/defaults.yaml")
		t.Setenv("INPUT_LOGLEVEL", "FOOBAR")
		_, err := New()
		if err == nil {
			t.Errorf("Expected an error for an invalid LogLevel")
		}
	})

	t.Run("Test empty loglevel", func(t *testing.T) {
		t.Setenv("INPUT_FILES", "../example/defaults.yaml")
		t.Setenv("INPUT_LOGLEVEL", "")
		plan, err := New()
		if err != nil {
			t.Logf("Got error %v", err)
			t.Errorf("Error when loading yaml files")
		}
		if plan.LogLevel != "WARN" {
			t.Errorf("Expected 'WARN' as default LogLevel, got %v", plan.LogLevel)
		}
	})
}
