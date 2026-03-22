package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var contractsDir string
	flag.StringVar(&contractsDir, "dir", "contracts", "Path to contracts directory")
	flag.Parse()

	var errs []string
	if err := filepath.WalkDir(contractsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: walk error: %v", path, err))
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			errs = append(errs, fmt.Sprintf("%s: read error: %v", path, readErr))
			return nil
		}

		switch {
		case strings.Contains(path, string(filepath.Separator)+"http"+string(filepath.Separator)):
			if e := validateHTTPContract(path, string(content)); e != "" {
				errs = append(errs, e)
			}
		case strings.Contains(path, string(filepath.Separator)+"events"+string(filepath.Separator)):
			if e := validateEventContract(path, string(content)); e != "" {
				errs = append(errs, e)
			}
		default:
			// Skip unknown groups.
		}
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "walk failed: %v\n", err)
		os.Exit(1)
	}

	if len(errs) > 0 {
		fmt.Println("contract validation failed:")
		for _, e := range errs {
			fmt.Printf("- %s\n", e)
		}
		os.Exit(1)
	}

	fmt.Printf("contracts validated: %s\n", contractsDir)
}

func validateHTTPContract(path, content string) string {
	required := []string{"openapi:", "info:", "paths:"}
	missing := missingKeys(required, content)
	if len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf("%s: missing required keys: %s", path, strings.Join(missing, ", "))
}

func validateEventContract(path, content string) string {
	// Support both old and new contract shapes.
	if strings.Contains(content, "kind: event-contract") || strings.Contains(content, "kind: event-envelope") {
		required := []string{"version:", "name:"}
		missing := missingKeys(required, content)
		if len(missing) == 0 {
			return ""
		}
		return fmt.Sprintf("%s: missing required keys: %s", path, strings.Join(missing, ", "))
	}

	// Legacy fallback check.
	required := []string{"version:", "topic:"}
	missing := missingKeys(required, content)
	if len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf("%s: unknown event contract shape, missing keys: %s", path, strings.Join(missing, ", "))
}

func missingKeys(required []string, content string) []string {
	index := make(map[string]struct{}, len(required))
	for _, k := range required {
		index[k] = struct{}{}
	}

	sc := bufio.NewScanner(strings.NewReader(content))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		for k := range index {
			if strings.HasPrefix(line, k) {
				delete(index, k)
			}
		}
	}

	if len(index) == 0 {
		return nil
	}
	missing := make([]string, 0, len(index))
	for k := range index {
		missing = append(missing, k)
	}
	return missing
}
