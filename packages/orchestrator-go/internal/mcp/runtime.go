package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	oconfig "github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/config"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/dispatch"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/types"
)

type RuntimeConfig struct {
	ExecutionMode types.ExecutionMode
	ConfigPath    string
	A2AConfig     *oconfig.Config
}

type RuntimeConfigOptions struct {
	ProjectRoot           string
	ConfigPath            string
	ConfigPathExplicit    bool
	ExecutionMode         string
	ExecutionModeExplicit bool
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{ExecutionMode: types.ExecutionNative}
}

func NewRuntimeConfig(mode string) (RuntimeConfig, error) {
	trimmed := strings.TrimSpace(mode)
	if trimmed == "" {
		return DefaultRuntimeConfig(), nil
	}

	switch types.ExecutionMode(trimmed) {
	case types.ExecutionNative, types.ExecutionA2A, types.ExecutionHybrid:
		return RuntimeConfig{ExecutionMode: types.ExecutionMode(trimmed)}, nil
	default:
		return RuntimeConfig{}, fmt.Errorf("invalid execution mode %q (expected native, a2a, or hybrid)", mode)
	}
}

func LoadRuntimeConfig(options RuntimeConfigOptions) (RuntimeConfig, error) {
	runtime := DefaultRuntimeConfig()

	selectedPath, explicitPath, err := selectConfigPath(options)
	if err != nil {
		return RuntimeConfig{}, err
	}
	if selectedPath != "" {
		loaded, err := oconfig.LoadFile(selectedPath)
		if err != nil {
			if explicitPath {
				return RuntimeConfig{}, fmt.Errorf("load orchestrator config: %w", err)
			}
			if !os.IsNotExist(err) {
				return RuntimeConfig{}, fmt.Errorf("load orchestrator config: %w", err)
			}
		} else {
			runtime.ConfigPath = selectedPath
			runtime.A2AConfig = loaded
			runtime.ExecutionMode = loaded.Execution.Mode
		}
	}

	if options.ExecutionModeExplicit {
		overridden, err := NewRuntimeConfig(options.ExecutionMode)
		if err != nil {
			return RuntimeConfig{}, err
		}
		runtime.ExecutionMode = overridden.ExecutionMode
	}

	return runtime, nil
}

func selectConfigPath(options RuntimeConfigOptions) (path string, explicit bool, err error) {
	if strings.TrimSpace(options.ProjectRoot) != "" {
		path = filepath.Join(options.ProjectRoot, oconfig.DefaultProjectConfigPath)
	}
	if envPath := strings.TrimSpace(os.Getenv("AI_SETUP_ORCHESTRATOR_CONFIG")); envPath != "" {
		path = envPath
		explicit = true
	}
	if options.ConfigPathExplicit || strings.TrimSpace(options.ConfigPath) != "" {
		if strings.TrimSpace(options.ConfigPath) == "" {
			return "", false, fmt.Errorf("--config requires a path")
		}
		path = options.ConfigPath
		explicit = true
	}
	if path == "" {
		return "", explicit, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", explicit, err
	}
	return abs, explicit, nil
}

type OrchestratorOption func(*Orchestrator)

func WithRuntimeConfig(config RuntimeConfig) OrchestratorOption {
	return func(o *Orchestrator) {
		o.Runtime = config
		o.Dispatcher = defaultDispatcherFor(config)
	}
}

func WithDispatcher(dispatcher dispatch.Dispatcher) OrchestratorOption {
	return func(o *Orchestrator) {
		if dispatcher != nil {
			o.Dispatcher = dispatcher
		}
	}
}

func defaultDispatcherFor(config RuntimeConfig) dispatch.Dispatcher {
	switch config.ExecutionMode {
	case types.ExecutionA2A:
		if config.A2AConfig != nil {
			return dispatch.NewConfiguredDispatcher(types.ExecutionA2A, config.A2AConfig)
		}
		return dispatch.NewConfiguredDispatcher(types.ExecutionA2A, nil)
	case types.ExecutionHybrid:
		return dispatch.NewConfiguredDispatcher(types.ExecutionHybrid, config.A2AConfig)
	case types.ExecutionNative:
		fallthrough
	default:
		return dispatch.NewNativeDispatcher()
	}
}
