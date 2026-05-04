package scout

import "testing"

func TestDetectFrameworksFromPackageJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "package.json", `{
  "dependencies": {
    "next": "13.5.1",
    "react": "^18.2.0",
    "@prisma/client": "5.0.0"
  },
  "devDependencies": {
    "vitest": "^1.2.3"
  }
}`)

	frameworks := frameworkEntriesByName(DetectFrameworks(dir))

	assertFramework(t, frameworks, "Next.js", "13.5.1", "package.json:dependencies.next")
	assertFramework(t, frameworks, "React", "^18.2.0", "package.json:dependencies.react")
	assertFramework(t, frameworks, "Prisma", "5.0.0", "package.json:dependencies.@prisma/client")
	assertFramework(t, frameworks, "Vitest", "^1.2.3", "package.json:devDependencies.vitest")
}

func TestDetectFrameworksFromGoModAndPython(t *testing.T) {
	tests := []struct {
		name string
		file string
		body string
		want string
	}{
		{
			name: "go module require",
			file: "go.mod",
			body: "module example.com/app\n\nrequire (\n\tgithub.com/gin-gonic/gin v1.9.1\n\tgorm.io/gorm v1.25.5\n)\n",
			want: "Gin",
		},
		{
			name: "requirements",
			file: "requirements.txt",
			body: "fastapi==0.110.0\nSQLAlchemy>=2.0\n",
			want: "FastAPI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeTestFile(t, dir, tt.file, tt.body)

			frameworks := frameworkEntriesByName(DetectFrameworks(dir))
			if _, ok := frameworks[tt.want]; !ok {
				t.Fatalf("framework %q not detected in %#v", tt.want, frameworks)
			}
		})
	}
}

func frameworkEntriesByName(entries []FrameworkEntry) map[string]FrameworkEntry {
	byName := make(map[string]FrameworkEntry, len(entries))
	for _, entry := range entries {
		byName[entry.Name] = entry
	}
	return byName
}

func assertFramework(t *testing.T, entries map[string]FrameworkEntry, name, version, source string) {
	t.Helper()
	entry, ok := entries[name]
	if !ok {
		t.Fatalf("framework %q not detected in %#v", name, entries)
	}
	if entry.Version != version || entry.Source != source {
		t.Fatalf("%s = %#v, want version %q source %q", name, entry, version, source)
	}
}
