package scout

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DetectInfra detects CI/CD and Docker-related files.
func DetectInfra(targetDir string) ([]string, *DockerInfo) {
	var ci []string
	if dirExists(filepath.Join(targetDir, ".github", "workflows")) {
		ci = append(ci, ".github/workflows/")
	}
	ciFiles := []string{".gitlab-ci.yml", "Jenkinsfile", ".travis.yml", "azure-pipelines.yml", "bitbucket-pipelines.yml", ".drone.yml", "buildspec.yml", "cloudbuild.yaml"}
	for _, rel := range ciFiles {
		if fileExists(filepath.Join(targetDir, rel)) {
			ci = append(ci, rel)
		}
	}
	if dirExists(filepath.Join(targetDir, ".circleci")) {
		ci = append(ci, ".circleci/")
	}
	sort.Strings(ci)

	var dockerfiles []string
	var composeFiles []string
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
		rel, err := filepath.Rel(targetDir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if name == "Dockerfile" || strings.HasPrefix(name, "Dockerfile.") {
			dockerfiles = append(dockerfiles, rel)
		}
		if name == "docker-compose.yml" || name == "docker-compose.yaml" {
			composeFiles = append(composeFiles, rel)
		}
		return nil
	})
	sort.Strings(dockerfiles)
	sort.Strings(composeFiles)
	if len(dockerfiles) == 0 && len(composeFiles) == 0 {
		return ci, nil
	}
	docker := &DockerInfo{}
	if len(dockerfiles) > 0 {
		docker.Dockerfile = dockerfiles[0]
	}
	if len(composeFiles) > 0 {
		docker.Compose = composeFiles[0]
	}
	return ci, docker
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
