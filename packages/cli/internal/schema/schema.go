package schema

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed *.json
var FS embed.FS

// LazyAI returns the embedded LazyAI manifest schema bytes.
func LazyAI() []byte {
	data, err := fs.ReadFile(FS, "lazyai.schema.json")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded lazyai schema: %v", err))
	}
	return data
}

// Lock returns the embedded lockfile schema bytes.
func Lock() []byte {
	data, err := fs.ReadFile(FS, "lock.schema.json")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded lock schema: %v", err))
	}
	return data
}

// MCPCatalog returns the embedded MCP catalog schema bytes.
func MCPCatalog() []byte {
	data, err := fs.ReadFile(FS, "mcp-catalog.schema.json")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded mcp catalog schema: %v", err))
	}
	return data
}

// EvalCase returns the embedded eval case schema bytes.
func EvalCase() []byte {
	data, err := fs.ReadFile(FS, "eval-case.schema.json")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded eval-case schema: %v", err))
	}
	return data
}

// EvalHoldout returns the embedded eval holdout schema bytes.
func EvalHoldout() []byte {
	data, err := fs.ReadFile(FS, "eval-holdout.schema.json")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded eval-holdout schema: %v", err))
	}
	return data
}

// EvalRubric returns the embedded eval rubric schema bytes.
func EvalRubric() []byte {
	data, err := fs.ReadFile(FS, "eval-rubric.schema.json")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded eval-rubric schema: %v", err))
	}
	return data
}

// Names returns the logical names of embedded schema artifacts.
func Names() []string {
	return []string{"lazyai", "lock", "mcp-catalog", "eval-case", "eval-holdout", "eval-rubric"}
}
