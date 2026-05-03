package scout

import "time"

var DefaultExcludedDirs = []string{
	"node_modules", ".git", ".reversa", "_reversa_sdd",
	"dist", "build", "coverage", "__pycache__", ".cache",
	"vendor", "tmp", "temp", ".next", ".nuxt", ".output",
	".worktrees", ".ai-setup.db",
}

// RunScout is the headless, deterministic Scout. It scans the target
// directory and returns SurfaceData. No AI involved — pure filesystem
// analysis following Reversa's documented Scout methodology.
func RunScout(targetDir string) (*SurfaceData, error) {
	data := &SurfaceData{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		ProjectRoot: targetDir,
	}

	data.Languages = DetectLanguages(targetDir, DefaultExcludedDirs)
	data.PrimaryLanguage = primaryLanguage(data.Languages)
	data.Frameworks = DetectFrameworks(targetDir)
	data.PackageManager = DetectPackageManager(targetDir)
	data.EntryPoints = DetectEntryPoints(targetDir)
	data.Modules = DetectModules(targetDir, DefaultExcludedDirs)
	data.ConfigFiles = detectConfigFiles(targetDir)
	data.CICD, data.Docker = DetectInfra(targetDir)
	data.DatabaseHints = DetectDatabaseHints(targetDir)
	data.TestFramework, data.TestFileCount = detectTestFramework(targetDir)
	data.TotalFiles = countTotalFiles(targetDir, DefaultExcludedDirs)

	return data, nil
}

func primaryLanguage(langs []LanguageEntry) string {
	if len(langs) == 0 {
		return ""
	}
	return langs[0].Name // sorted by file count descending
}

func excludedDirSet(excludedDirs []string) map[string]struct{} {
	set := make(map[string]struct{}, len(excludedDirs))
	for _, dir := range excludedDirs {
		set[dir] = struct{}{}
	}
	return set
}

func isExcludedDir(name string, excluded map[string]struct{}) bool {
	_, ok := excluded[name]
	return ok
}
