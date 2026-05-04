package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

const (
	DefaultProjectConfigPath = ".ai/orchestrator.json"

	AuthNone      = "none"
	AuthBearerEnv = "bearerEnv"
	AuthHeaderEnv = "headerEnv"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type Config struct {
	Version   int                       `json:"version"`
	Execution ExecutionConfig           `json:"execution"`
	Providers map[string]ProviderConfig `json:"providers,omitempty"`
	Agents    map[string]AgentConfig    `json:"agents,omitempty"`
}

type ExecutionConfig struct {
	Mode types.ExecutionMode `json:"mode"`
}

type ProviderConfig struct {
	Endpoint string     `json:"endpoint,omitempty"`
	CardURL  string     `json:"cardUrl,omitempty"`
	Auth     AuthConfig `json:"auth,omitempty"`
}

type AuthConfig struct {
	Type   string `json:"type"`
	Env    string `json:"env,omitempty"`
	Header string `json:"header,omitempty"`
}

type AgentConfig struct {
	Provider string   `json:"provider"`
	Enabled  *bool    `json:"enabled"`
	Tools    []string `json:"tools,omitempty"`
}

func (a AgentConfig) IsEnabled() bool {
	return a.Enabled != nil && *a.Enabled
}

func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return config, nil
}

func Parse(data []byte) (*Config, error) {
	if err := rejectInlineSecrets(data); err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("parse orchestrator config: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return nil, fmt.Errorf("parse orchestrator config: expected one JSON object")
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if c.Version != 1 {
		return fmt.Errorf("invalid config version %d (expected 1)", c.Version)
	}
	if c.Execution.Mode == "" {
		c.Execution.Mode = types.ExecutionNative
	}
	switch c.Execution.Mode {
	case types.ExecutionNative, types.ExecutionA2A, types.ExecutionHybrid:
	default:
		return fmt.Errorf("invalid execution mode %q (expected native, a2a, or hybrid)", c.Execution.Mode)
	}

	for name, provider := range c.Providers {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("provider name must not be empty")
		}
		if strings.TrimSpace(provider.Endpoint) == "" && strings.TrimSpace(provider.CardURL) == "" {
			return fmt.Errorf("provider %q must define endpoint or cardUrl", name)
		}
		if provider.Endpoint != "" {
			if err := validateHTTPURL(provider.Endpoint); err != nil {
				return fmt.Errorf("provider %q endpoint: %w", name, err)
			}
		}
		if provider.CardURL != "" {
			if err := validateHTTPURL(provider.CardURL); err != nil {
				return fmt.Errorf("provider %q cardUrl: %w", name, err)
			}
		}
		if err := validateAuth(name, provider.Auth); err != nil {
			return err
		}
	}

	for name, agent := range c.Agents {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("agent name must not be empty")
		}
		if agent.Enabled == nil {
			return fmt.Errorf("agent %q must define enabled", name)
		}
		if strings.TrimSpace(agent.Provider) == "" {
			return fmt.Errorf("agent %q must reference a provider", name)
		}
		if _, ok := c.Providers[agent.Provider]; !ok {
			return fmt.Errorf("agent %q references unknown provider %q", name, agent.Provider)
		}
		for _, tool := range agent.Tools {
			switch tool {
			case string(types.HostOpenCode), string(types.HostClaudeCode), string(types.HostCopilot):
			default:
				return fmt.Errorf("agent %q has unsupported tool %q (expected opencode, claude-code, or copilot)", name, tool)
			}
		}
	}

	return nil
}

func validateHTTPURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("must use http or https URL")
	}
	if parsed.Host == "" {
		return fmt.Errorf("must include host")
	}
	for key := range parsed.Query() {
		if isSecretLikeKey(key) {
			return fmt.Errorf("must not include secret-like query parameter %q", key)
		}
	}
	return nil
}

func validateAuth(providerName string, auth AuthConfig) error {
	authType := auth.Type
	if authType == "" {
		authType = AuthNone
	}
	switch authType {
	case AuthNone:
		if auth.Env != "" || auth.Header != "" {
			return fmt.Errorf("provider %q auth type none must not define env or header", providerName)
		}
	case AuthBearerEnv:
		if err := validateEnvRef(providerName, auth.Env); err != nil {
			return err
		}
		if auth.Header != "" {
			return fmt.Errorf("provider %q bearerEnv auth must not define header", providerName)
		}
	case AuthHeaderEnv:
		if err := validateEnvRef(providerName, auth.Env); err != nil {
			return err
		}
		if strings.TrimSpace(auth.Header) == "" {
			return fmt.Errorf("provider %q headerEnv auth must define header", providerName)
		}
	default:
		return fmt.Errorf("provider %q has unsupported auth type %q (expected none, bearerEnv, or headerEnv)", providerName, auth.Type)
	}
	return nil
}

func validateEnvRef(providerName, env string) error {
	if strings.TrimSpace(env) == "" {
		return fmt.Errorf("provider %q auth must define env", providerName)
	}
	if !envNamePattern.MatchString(env) {
		return fmt.Errorf("provider %q auth env %q is not a valid environment variable reference", providerName, env)
	}
	return nil
}

func rejectInlineSecrets(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	return rejectInlineSecretsValue(raw, "")
}

func rejectInlineSecretsValue(value any, path string) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			nextPath := key
			if path != "" {
				nextPath = path + "." + key
			}
			if isSecretLikeKey(key) {
				return fmt.Errorf("inline secret-like field %q is not allowed; use env references", nextPath)
			}
			if err := rejectInlineSecretsValue(nested, nextPath); err != nil {
				return err
			}
		}
	case []any:
		for _, nested := range typed {
			if err := rejectInlineSecretsValue(nested, path); err != nil {
				return err
			}
		}
	}
	return nil
}

func isSecretLikeKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(key, "_", ""))
	for _, marker := range []string{"secret", "token", "password", "apikey", "credential"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}
