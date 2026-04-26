package setupscan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/adapter"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/frontmatter"
)

var reusableAgentIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

type ObservedAgent struct {
	ID          string            `json:"id"`
	Directory   string            `json:"directory"`
	PromptPath  string            `json:"promptPath,omitempty"`
	Status      string            `json:"status"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Tools       []string          `json:"tools,omitempty"`
	MCP         *ObservedAgentMCP `json:"mcp,omitempty"`
	Reasons     []string          `json:"reasons,omitempty"`
	PromptBody  string            `json:"-"`
}

type ObservedAgentMCP struct {
	ConfigPath  string   `json:"configPath"`
	Scoped      bool     `json:"scoped"`
	ServerNames []string `json:"serverNames,omitempty"`
	ServerCount int      `json:"serverCount"`
}

type agentMCPConfig struct {
	MCPServers map[string]adapter.McpServer `json:"mcpServers"`
}

func observedAgents(opts Options) []ObservedAgent {
	agentsRoot := filepath.Join(opts.TargetDir, ".ai", "agents")
	if !files.DirExists(agentsRoot) {
		return nil
	}

	entries, err := os.ReadDir(agentsRoot)
	if err != nil {
		return []ObservedAgent{invalidObservedAgent(".ai/agents", agentsRoot, "invalid-agents-root:"+err.Error())}
	}

	agents := make([]ObservedAgent, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agents = append(agents, observeAgentDir(filepath.Join(agentsRoot, entry.Name()), entry.Name()))
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].ID < agents[j].ID })
	return agents
}

func observeAgentDir(dirPath, dirName string) ObservedAgent {
	agent := ObservedAgent{
		ID:         dirName,
		Directory:  dirPath,
		PromptPath: filepath.Join(dirPath, "AGENT.md"),
		Status:     "detected",
	}

	if !reusableAgentIDPattern.MatchString(dirName) {
		agent.Status = "invalid"
		agent.Reasons = append(agent.Reasons, "invalid-agent-id")
	}

	if err := hydrateObservedAgent(&agent); err != nil {
		agent.Status = "invalid"
		agent.Reasons = append(agent.Reasons, err.Error())
	}

	sort.Strings(agent.Tools)
	sort.Strings(agent.Reasons)
	return agent
}

func hydrateObservedAgent(agent *ObservedAgent) error {
	data, err := os.ReadFile(agent.PromptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("missing-agent-md")
		}
		return fmt.Errorf("read-agent-md:%w", err)
	}

	fm, body, err := frontmatter.ParseYamlFrontmatter(string(data))
	if err != nil {
		return fmt.Errorf("invalid-agent-frontmatter:%w", err)
	}
	agent.PromptBody = strings.TrimSpace(body)
	if agent.PromptBody == "" {
		return fmt.Errorf("empty-agent-body")
	}
	agent.Title = firstNonEmpty(stringField(fm, "title"), stringField(fm, "name"), extractFirstHeading(agent.PromptBody))
	agent.Description = firstNonEmpty(stringField(fm, "description"), extractFirstParagraph(agent.PromptBody))
	agent.Tools = normalizeStringList(fm, "tools")

	mcpPath := filepath.Join(agent.Directory, "mcp.json")
	if !files.FileExists(mcpPath) {
		return nil
	}
	mcp, err := parseObservedAgentMCP(mcpPath)
	if err != nil {
		return err
	}
	agent.MCP = mcp
	return nil
}

func parseObservedAgentMCP(path string) (*ObservedAgentMCP, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read-agent-mcp:%w", err)
	}

	var topLevel map[string]json.RawMessage
	if err := json.Unmarshal(data, &topLevel); err != nil {
		return nil, fmt.Errorf("invalid-agent-mcp-json:%w", err)
	}
	if len(topLevel) != 1 {
		return nil, fmt.Errorf("invalid-agent-mcp-schema")
	}
	rawServers, ok := topLevel["mcpServers"]
	if !ok {
		return nil, fmt.Errorf("invalid-agent-mcp-schema")
	}

	var config agentMCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid-agent-mcp-schema:%w", err)
	}
	if string(rawServers) == "null" || config.MCPServers == nil {
		return nil, fmt.Errorf("invalid-agent-mcp-schema")
	}

	serverNames := make([]string, 0, len(config.MCPServers))
	for name := range config.MCPServers {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return nil, fmt.Errorf("invalid-agent-mcp-server-name")
		}
		serverNames = append(serverNames, trimmed)
	}
	sort.Strings(serverNames)

	return &ObservedAgentMCP{
		ConfigPath:  path,
		Scoped:      true,
		ServerNames: serverNames,
		ServerCount: len(serverNames),
	}, nil
}

func invalidObservedAgent(id, dir string, reason string) ObservedAgent {
	return ObservedAgent{ID: id, Directory: dir, Status: "invalid", Reasons: []string{reason}}
}

func stringField(fm map[string]any, key string) string {
	if fm == nil {
		return ""
	}
	value, _ := fm[key].(string)
	return strings.TrimSpace(value)
}

func normalizeStringList(fm map[string]any, key string) []string {
	if fm == nil {
		return nil
	}
	value, ok := fm[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case string:
		if trimmed := strings.TrimSpace(typed); trimmed != "" {
			return []string{trimmed}
		}
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			str, ok := item.(string)
			if !ok {
				continue
			}
			if trimmed := strings.TrimSpace(str); trimmed != "" {
				values = append(values, trimmed)
			}
		}
		return values
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func extractFirstHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}
	return ""
}

func extractFirstParagraph(content string) string {
	paragraphs := strings.Split(content, "\n\n")
	for _, paragraph := range paragraphs {
		trimmed := strings.TrimSpace(paragraph)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		return strings.Join(strings.Fields(trimmed), " ")
	}
	return ""
}
