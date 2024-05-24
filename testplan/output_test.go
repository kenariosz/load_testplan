package testplan

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"testing"
)

func TestOutput(t *testing.T) {
	if os.Getenv("GITHUB_ENV") == "" {
		// running locally
		os.Setenv("GITHUB_ENV", "../env.tmp")
	}
	t.Setenv("INPUT_FILES", "../example/defaults.yaml")
	t.Setenv("GITHUB_OUTPUT", "../output.tmp")
	t.Setenv("INPUT_SET_OUTPUT", "true")
	t.Setenv("GITHUB_ENV", "../env.tmp")
	t.Setenv("INPUT_SET_ENV", "true")
	t.Setenv("INPUT_SET_PRINT", "true")
	t.Setenv("INPUT_YAML", "../yaml.tmp")
	t.Setenv("INPUT_JSON", "../json.tmp")
	t.Setenv("INPUT_GENERATE_JOB", "true")
	t.Setenv("INPUT_LOGLEVEL", "TRACE")

	plan, err := New()
	if err != nil {
		t.Logf("Got error %v", err)
		t.Errorf("Error when loading yaml file")
	}
	err = plan.Output()
	if err != nil {
		t.Errorf("Error generating the output")
	}

	t.Run("Test output", func(t *testing.T) {
		pattern := "(?m:^string<<_GitHubActionsFileCommandDelimeter_\r*\nString\r*\n_GitHubActionsFileCommandDelimeter_\r*\n)"
		r := regexp.MustCompile(pattern)
		b, err := os.ReadFile("../output.tmp")
		if err != nil {
			t.Errorf("Could not read file ../output.tmp")
		}
		s := string(b)
		found := r.FindAllStringSubmatch(s, -1)
		if found == nil {
			t.Errorf("Did not find output for variable 'string' with regexp search: %v", found)
		}
	})

	t.Run("Test env", func(t *testing.T) {
		pattern := "(?m:^string<<_GitHubActionsFileCommandDelimeter_\r*\nString\r*\n_GitHubActionsFileCommandDelimeter_\r*\n)"
		r := regexp.MustCompile(pattern)
		b, err := os.ReadFile("../env.tmp")
		if err != nil {
			t.Errorf("Could not read file ../env.tmp")
		}
		s := string(b)
		found := r.FindAllStringSubmatch(s, -1)
		if found == nil {
			t.Errorf("Did not find output for variable 'string' with regexp search: %v", found)
		}
	})

	t.Run("Test yaml", func(t *testing.T) {
		b, err := os.ReadFile("../yaml.tmp")
		if err != nil {
			t.Errorf("Could not read '../yaml.tmp', reason: %v", err)
		}
		var data map[string]interface{}
		if err := yaml.Unmarshal(b, &data); err != nil {
			t.Errorf("Could not unmarshall yaml, reason: %v", err)
		}
		if data["string"].(string) != "String" {
			t.Errorf("Field 'string' does not contain 'String'")
		}
	})

	t.Run("Test json", func(t *testing.T) {
		b, err := os.ReadFile("../json.tmp")
		if err != nil {
			t.Errorf("Could not read '../json.tmp', reason: %v", err)
		}
		var data map[string]interface{}
		if err := json.Unmarshal(b, &data); err != nil {
			t.Errorf("Could not unmarshall json, reason: %v", err)
		}
		if data["string"].(string) != "String" {
			t.Errorf("Field 'string' does not contain 'String'")
		}
	})

	t.Run("Test job", func(t *testing.T) {
		b, err := os.ReadFile("job_load_testplan.yml")
		if err != nil {
			t.Errorf("Could not read 'job_load_testplan.yml', reason: %v", err)
		}
		var data map[string]interface{}
		if err := yaml.Unmarshal(b, &data); err != nil {
			t.Errorf("Could not unmarshall yaml, reason: %v, data %v ", err, string(b))
		}
		j, ok := data["jobs"].(map[string]interface{})
		if !ok {
			t.Errorf("Could not find key 'jobs'")
		}
		l, ok := j["load_testplan"].(map[string]interface{})
		if !ok {
			t.Errorf("Could not find key 'jobs.load_testplan'")
		}
		o, ok := l["outputs"].(map[string]interface{})
		if !ok {
			t.Errorf("Could not find key 'jobs.load_testplan.outputs'")
		}
		s, ok := o["string"].(string)
		if !ok {
			t.Errorf("Could not find key 'jobs.load_testplan.outputs.string'")
		}
		if s != "${{ steps.ltp.outputs.string }}" {
			t.Errorf("'jobs.load_testplan.outputs.string' does not contain '${{ steps.ltp.outputs.string }}' but '%v'", s)
		}
	})
}

func TestOutputImpossible(t *testing.T) {
	if os.Getenv("GITHUB_ENV") == "" {
		// running locally
		os.Setenv("GITHUB_ENV", "../env.tmp")
	}
	t.Setenv("INPUT_FILES", "../example/defaults.yaml")
	t.Setenv("INPUT_LOGLEVEL", "TRACE")

	t.Run("Non-existing folder for yaml output", func(t *testing.T) {
		t.Setenv("INPUT_YAML", "this_folder_should_not_exist/yaml.tmp")
		plan, err := New()
		if err != nil {
			t.Logf("Got error %v", err)
			t.Errorf("Error when loading yaml file")
		}
		err = plan.Output()
		if err != nil {
			t.Errorf("Got error ' %v', when the folder for the yaml output doesn't exist.", err)
		}
	})

	t.Run("Non-existing folder for json output", func(t *testing.T) {
		t.Setenv("INPUT_JSON", "this_folder_should_not_exist/json.tmp")
		plan, err := New()
		if err != nil {
			t.Logf("Got error %v", err)
			t.Errorf("Error when loading yaml file")
		}
		err = plan.Output()
		if err != nil {
			t.Errorf("Got error ' %v', when the folder for the yaml output doesn't exist.", err)
		}
	})
}
