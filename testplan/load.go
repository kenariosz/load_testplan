package testplan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"
)

// Load all files as templates, convert them into maps and merge them together
func (plan *Testplan) loadFiles() error {
	logger := log.With().Str("func", "loadFiles").Str("package", "testplan").Logger()
	logger.Trace().Msg("Enter func")
	for _, f := range plan.Files {
		log.Debug().Str("file", f).Msg("Load file")

		input, err := plan.parseFile(f)
		if err != nil {
			return err
		}

		var data map[string]interface{}
		t := plan.InputType
		if t == "auto" {
			s := strings.Split(f, ".")
			suffix := s[len(s)-1]
			switch suffix {
			case "json", "jso", "jsn", "js":
				t = "json"
			case "yaml", "yml":
				t = "yaml"
			default:
				err := errors.New("Unkonwn file suffix " + suffix + ". Can't use input type 'auto'.")
				log.Error().Err(err).Str("file", f).Str("suffix", suffix).Msg("Can't determine file type")
				fmt.Println("::error :: Error: Unkonwn file suffix " + suffix + ". Can't use input type 'auto'.")
				return err
			}
		}

		switch t {
		case "json":
			err = json.Unmarshal(input, &data)
			if err != nil {
				log.Error().Err(err).Str("file", f).Msg("Could not unmarshall json")
				fmt.Println("::error :: Error: Could not unmarshall json:" + err.Error())
				return err
			}
		default:
			err = yaml.Unmarshal(input, &data)
			if err != nil {
				log.Error().Err(err).Str("file", f).Msg("Could not unmarshall yaml")
				fmt.Println("::error :: Error: Could not unmarshall json:" + err.Error())
				return err
			}
		}
		plan.Data = mergeMaps(plan.Data, data)
		if plan.LogLevel == "TRACE" {
			spew.Dump(plan.Data)
		}
	}
	return nil
}

// get a file content from an URL
func getFromURL(url string) (string, error) {
	logger := log.With().Str("func", "getFromURL").Str("package", "testplan").Str("url", url).Logger()
	logger.Trace().Msg("Enter func")

	r, err := http.Get(url)
	if err != nil {
		logger.Error().Err(err).Msg("Get from URL failed")
		fmt.Println("::error :: Error: Get from URL failed:" + err.Error())
		return "", err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		e := fmt.Errorf("HTTP status %v is not OK", r.StatusCode)
		logger.Error().Err(e).Msg("Unsupported HTTP status")
		fmt.Println("::error :: Error: " + e.Error())
		return "", e
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to read HTTP response body")
		fmt.Println("::error :: Error: Failed to read HTTP response body:" + err.Error())
		return "", err
	}

	return string(data), nil
}

// Uses text/template templating when loading the yaml files
func (plan *Testplan) parseFile(name string) ([]byte, error) {
	logger := log.With().Str("func", "readFile").Str("package", "testplan").Logger()
	logger.Trace().Msg("Enter func")
	var raw string
	var b bytes.Buffer
	var t *template.Template
	var template_name string

	u, err := url.ParseRequestURI(name)
	if err != nil || u.Scheme == "file" {
		s, err := os.ReadFile(name)
		template_name = path.Base(name)
		raw = string(s)
		if err != nil {
			log.Error().Err(err).Str("file", name).Msg("Failed to read file")
			fmt.Println("::error :: Error: Failed to read file '" + name + "':" + err.Error())
			return b.Bytes(), err
		}
	} else {
		template_name = path.Base(u.Path)
		n := name
		if plan.Token != "" {
			n = n + "?token=" + plan.Token
			log.Debug().Msg("Adding token to url")
		}
		raw, err = getFromURL(n)
		if err != nil {
			return b.Bytes(), err
		}
	}
	t = template.New(template_name)
	t, err = t.Parse(raw)
	if err != nil {
		log.Error().Err(err).Str("file", name).Msg("Failed to parse template")
		fmt.Println("::error :: Error: Failed to parse template:" + err.Error())
		return b.Bytes(), err
	}
	if err = t.Execute(&b, plan); err != nil {
		log.Error().Err(err).Str("file", name).Msg("Failed to execute template")
		fmt.Println("::error :: Error: Failed to execute template:" + err.Error())
		return b.Bytes(), err
	}
	log.Trace().Str("raw", raw).Str("parsed", string(b.Bytes())).Msg("Parsed file")
	return b.Bytes(), nil
}

// Merge two maps
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
