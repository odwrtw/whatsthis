package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type rawCase struct {
	Input string
	Value any
}

type outCase struct {
	FName  string         `json:"fname"`
	Wanted map[string]any `json:"wanted"`
	Src    string         `json:"src"`
	Opts   map[string]any `json:"opts,omitempty"`
}

func main() {
	inDir := flag.String("in", "original-guessit/guessit/test", "input yaml dir")
	outDir := flag.String("out", "testdata/guessit", "output json dir")
	flag.Parse()

	files := []string{"episodes.yml", "movies.yml", "various.yml"}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fail(err)
	}

	for _, name := range files {
		if err := convertOne(filepath.Join(*inDir, name), filepath.Join(*outDir, strings.TrimSuffix(name, ".yml")+".tests.json")); err != nil {
			fail(err)
		}
	}
}

func convertOne(inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", inPath, err)
	}

	var root yaml.MapSlice
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("yaml %s: %w", inPath, err)
	}

	defaults := map[string]any{}
	var raws []rawCase
	for _, item := range root {
		key := strings.TrimSpace(fmt.Sprint(item.Key))
		if key == "__default__" {
			d, err := decodeMap(item.Value)
			if err != nil {
				return fmt.Errorf("defaults %s: %w", inPath, err)
			}
			defaults = d
			continue
		}
		raws = append(raws, rawCase{Input: key, Value: item.Value})
	}

	resolved := make([]map[string]any, len(raws))
	var nextExpected map[string]any
	for i := len(raws) - 1; i >= 0; i-- {
		if raws[i].Value == nil {
			resolved[i] = cloneMap(nextExpected)
			continue
		}
		d, err := decodeMap(raws[i].Value)
		if err != nil {
			return fmt.Errorf("case %s[%d]: %w", inPath, i, err)
		}
		resolved[i] = d
		nextExpected = cloneMap(d)
	}

	source := filepath.Base(inPath)
	outItems := make([]outCase, 0, len(raws))
	for i, rc := range raws {
		exp := resolved[i]
		applyDefaults(exp, defaults)
		input := filepath.Base(strings.ReplaceAll(rc.Input, "\\", "/"))
		opts := extractOptions(exp)
		out := outCase{
			FName:  input,
			Wanted: exp,
			Src:    source,
			Opts:   opts,
		}
		outItems = append(outItems, out)
		_ = i
	}

	payload, err := json.MarshalIndent(outItems, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", inPath, err)
	}
	payload = append(payload, '\n')
	if err := os.WriteFile(outPath, payload, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}

func extractOptions(exp map[string]any) map[string]any {
	v, ok := exp["options"]
	if !ok {
		return nil
	}
	opts, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	delete(exp, "options")
	return opts
}

func decodeMap(v any) (map[string]any, error) {
	if v == nil {
		return map[string]any{}, nil
	}
	switch x := v.(type) {
	case yaml.MapSlice:
		m := make(map[string]any, len(x))
		for _, item := range x {
			val, err := normalizeYAMLValue(item.Value)
			if err != nil {
				return nil, err
			}
			m[fmt.Sprint(item.Key)] = val
		}
		return m, nil
	case map[any]any:
		m := make(map[string]any, len(x))
		for k, v2 := range x {
			val, err := normalizeYAMLValue(v2)
			if err != nil {
				return nil, err
			}
			m[fmt.Sprint(k)] = val
		}
		return m, nil
	case map[string]any:
		return x, nil
	default:
		return map[string]any{}, nil
	}
}

func normalizeYAMLValue(v any) (any, error) {
	switch x := v.(type) {
	case yaml.MapSlice:
		return decodeMap(x)
	case []any:
		out := make([]any, 0, len(x))
		for _, e := range x {
			n, err := normalizeYAMLValue(e)
			if err != nil {
				return nil, err
			}
			out = append(out, n)
		}
		return out, nil
	default:
		return v, nil
	}
}

func applyDefaults(dst, defaults map[string]any) {
	for k, v := range defaults {
		if _, ok := dst[k]; !ok {
			dst[k] = v
		}
	}
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
