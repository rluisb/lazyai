package scout

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// detectTestFramework detects the most likely test framework and counts test files.
func detectTestFramework(targetDir string) (string, int) {
	framework := ""
	count := 0
	excluded := excludedDirSet(DefaultExcludedDirs)
	_ = filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != targetDir && isExcludedDir(d.Name(), excluded) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		name := d.Name()
		if framework == "" {
			framework = frameworkFromTestConfig(name)
		}
		if isTestFile(name) {
			count++
			if framework == "" {
				framework = frameworkFromTestFile(name)
			}
		}
		return nil
	})
	return framework, count
}

func frameworkFromTestConfig(name string) string {
	switch {
	case strings.HasPrefix(name, "jest.config."):
		return "Jest"
	case strings.HasPrefix(name, "vitest.config."):
		return "Vitest"
	case strings.HasPrefix(name, ".mocharc."):
		return "Mocha"
	case strings.HasPrefix(name, "playwright.config."):
		return "Playwright"
	case strings.HasPrefix(name, "cypress.config."):
		return "Cypress"
	case strings.HasPrefix(name, "karma.conf."):
		return "Karma"
	case name == "pytest.ini" || name == "tox.ini":
		return "Pytest"
	default:
		return ""
	}
}

func isTestFile(name string) bool {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, "_test.go"):
		return true
	case jsTestName(lower):
		return true
	case strings.HasPrefix(lower, "test_") && strings.HasSuffix(lower, ".py"):
		return true
	case strings.HasSuffix(lower, "_test.py"):
		return true
	case strings.HasSuffix(lower, ".test.rb") || strings.HasSuffix(lower, "_spec.rb"):
		return true
	default:
		return false
	}
}

func jsTestName(name string) bool {
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx"} {
		if strings.HasSuffix(name, ".test"+ext) || strings.HasSuffix(name, ".spec"+ext) {
			return true
		}
	}
	return false
}

func frameworkFromTestFile(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, "_test.go"):
		return "Go test"
	case strings.HasPrefix(lower, "test_") && strings.HasSuffix(lower, ".py"), strings.HasSuffix(lower, "_test.py"):
		return "Pytest"
	case strings.HasSuffix(lower, ".test.rb") || strings.HasSuffix(lower, "_spec.rb"):
		return "RSpec"
	case jsTestName(lower):
		return "JavaScript test"
	default:
		return ""
	}
}
