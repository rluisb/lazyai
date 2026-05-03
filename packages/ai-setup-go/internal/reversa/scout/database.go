package scout

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DetectDatabaseHints detects ORM schemas, migration directories, and schema SQL files.
func DetectDatabaseHints(targetDir string) []DBHint {
	var hints []DBHint
	seen := map[string]struct{}{}
	add := func(path, typ string) {
		path = filepath.ToSlash(path)
		key := path + "\x00" + typ
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		hints = append(hints, DBHint{Path: path, Type: typ})
	}

	if fileExists(filepath.Join(targetDir, "prisma", "schema.prisma")) {
		add("prisma/schema.prisma", "prisma_schema")
	}
	if dirExists(filepath.Join(targetDir, "prisma", "migrations")) {
		add("prisma/migrations", "migrations_dir")
	}

	excluded := excludedDirSet(DefaultExcludedDirs)
	_ = filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != targetDir && isExcludedDir(d.Name(), excluded) {
				return filepath.SkipDir
			}
			rel, relErr := filepath.Rel(targetDir, path)
			if relErr == nil {
				rel = filepath.ToSlash(rel)
				switch {
				case d.Name() == "migrations":
					add(rel, "migrations_dir")
				case rel == "db/migrate":
					add(rel, "rails_migrations")
				case d.Name() == "alembic":
					add(rel, "alembic")
				}
			}
			return nil
		}
		if !d.Type().IsRegular() || strings.ToLower(filepath.Ext(d.Name())) != ".sql" {
			return nil
		}
		if sqlFileHasSchemaDDL(path) {
			rel, relErr := filepath.Rel(targetDir, path)
			if relErr == nil {
				add(rel, "sql_schema")
			}
		}
		return nil
	})

	priority := map[string]int{"prisma_schema": 0, "migrations_dir": 1, "rails_migrations": 2, "alembic": 2, "sql_schema": 3}
	sort.Slice(hints, func(i, j int) bool {
		pi, pj := priority[hints[i].Type], priority[hints[j].Type]
		if pi == pj {
			return hints[i].Path < hints[j].Path
		}
		return pi < pj
	})
	return hints
}

func sqlFileHasSchemaDDL(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 500)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	upper := bytes.ToUpper(buf[:n])
	return bytes.Contains(upper, []byte("CREATE TABLE")) || bytes.Contains(upper, []byte("ALTER TABLE"))
}
