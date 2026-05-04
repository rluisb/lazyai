package scout

type LanguageEntry struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
	FileCount  int      `json:"file_count"`
}

type FrameworkEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

type EntryPoint struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type DBHint struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type DockerInfo struct {
	Dockerfile string `json:"dockerfile,omitempty"`
	Compose    string `json:"compose,omitempty"`
}

type SurfaceData struct {
	GeneratedAt     string           `json:"generated_at"`
	ProjectRoot     string           `json:"project_root"`
	Languages       []LanguageEntry  `json:"languages"`
	PrimaryLanguage string           `json:"primary_language"`
	Frameworks      []FrameworkEntry `json:"frameworks"`
	PackageManager  string           `json:"package_manager"`
	EntryPoints     []EntryPoint     `json:"entry_points"`
	ConfigFiles     []string         `json:"config_files"`
	CICD            []string         `json:"ci_cd,omitempty"`
	Docker          *DockerInfo      `json:"docker,omitempty"`
	DatabaseHints   []DBHint         `json:"database_hints,omitempty"`
	TestFramework   string           `json:"test_framework"`
	TestFileCount   int              `json:"test_file_count"`
	Modules         []string         `json:"modules"`
	TotalFiles      int              `json:"total_files"`
}
