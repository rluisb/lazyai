package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	charmlog "charm.land/log/v2"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/dashboard"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	orchlog "github.com/rluisb/lazyai/packages/orchestrator/internal/log"
	orchmcp "github.com/rluisb/lazyai/packages/orchestrator/internal/mcp"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/queue"
)

const (
	defaultPort        = 57372
	defaultIdleTimeout = 10 * time.Minute
)

var (
	port        int
	projectRoot string
	execMode    string
	configPath  string
	scope       string
	globalRoot  string
	detachFlag  bool
	idleTimeout time.Duration
	logLevel    string
	logFormat   string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		daemonLogger().Error("command failed", "error", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "lazyai-orchestrator",
	Short: "Multi-agent orchestration runtime for LazyAI",
	Long:  `Coordinates chains, teams, and workflows with durable SQLite state and MCP tools.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return applyLoggingEnv(loggingFlagConfigFromCommand(cmd))
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStdio(cmd)
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the orchestrator as a shared HTTP daemon",
	Long: `Starts the orchestrator over HTTP/SSE for multi-client concurrent access.
Multiple AI CLI tools (Claude Code, OpenCode, Copilot) can connect simultaneously.

On startup a discovery file is written so other processes can detect the instance.
The server shuts down automatically after all SSE clients disconnect.`,
	RunE: runServe,
}

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to orchestrator daemon (auto-starts if needed)",
	Long: `Ensures the orchestrator daemon is running and creates a stdio MCP bridge.
This is the command AI CLI tools should use in their MCP config.

If the daemon is already running, just creates the bridge.
If no daemon is running, starts one in the background.`,
	RunE: runConnect,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Smart start (discover existing daemon or launch one)",
	RunE:  runStart,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Gracefully stop the orchestrator daemon",
	RunE:  runStop,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon health and active runs",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(serveCmd, connectCmd, startCmd, stopCmd, statusCmd)
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "Set log level (debug|info|warn|error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "Set log format (text|json|logfmt)")

	rootCmd.Flags().StringVar(&projectRoot, "project", "", "Project root path")
	rootCmd.Flags().StringVar(&scope, "scope", "project", "Scope: project|global|workspace")
	rootCmd.Flags().StringVar(&execMode, "execution-mode", "native", "Execution: native|a2a|hybrid")
	rootCmd.Flags().StringVar(&configPath, "config", "", "Orchestrator config path (default: .ai/orchestrator.json; env: AI_SETUP_ORCHESTRATOR_CONFIG)")

	serveCmd.Flags().IntVar(&port, "port", defaultPort, "TCP port")
	serveCmd.Flags().StringVar(&projectRoot, "project", "", "Project root path")
	serveCmd.Flags().StringVar(&scope, "scope", "project", "Scope: project|global|workspace")
	serveCmd.Flags().StringVar(&execMode, "execution-mode", "native", "Execution: native|a2a|hybrid")
	serveCmd.Flags().StringVar(&configPath, "config", "", "Orchestrator config path (default: .ai/orchestrator.json; env: AI_SETUP_ORCHESTRATOR_CONFIG)")
	serveCmd.Flags().DurationVar(&idleTimeout, "idle-timeout", defaultIdleTimeout, "Auto-shutdown idle timeout (0 disables, env: AI_SETUP_ORCHESTRATOR_IDLE_TIMEOUT)")
	serveCmd.Flags().BoolVar(&detachFlag, "detach", false, "Run in background")

	connectCmd.Flags().IntVar(&port, "port", defaultPort, "TCP port")
	connectCmd.Flags().StringVar(&projectRoot, "project", "", "Project root")
	connectCmd.Flags().StringVar(&scope, "scope", "project", "Scope")
	connectCmd.Flags().StringVar(&execMode, "execution-mode", "native", "Execution mode")
	connectCmd.Flags().StringVar(&configPath, "config", "", "Orchestrator config path")
	connectCmd.Flags().DurationVar(&idleTimeout, "idle-timeout", defaultIdleTimeout, "Auto-shutdown idle timeout for auto-started daemon (0 disables, env: AI_SETUP_ORCHESTRATOR_IDLE_TIMEOUT)")

	startCmd.Flags().IntVar(&port, "port", defaultPort, "TCP port")
	startCmd.Flags().StringVar(&projectRoot, "project", "", "Project root")
	startCmd.Flags().StringVar(&scope, "scope", "project", "Scope")
	startCmd.Flags().StringVar(&execMode, "execution-mode", "native", "Execution mode")
	startCmd.Flags().StringVar(&configPath, "config", "", "Orchestrator config path")
	startCmd.Flags().DurationVar(&idleTimeout, "idle-timeout", defaultIdleTimeout, "Auto-shutdown idle timeout (0 disables, env: AI_SETUP_ORCHESTRATOR_IDLE_TIMEOUT)")
}

type loggingFlagConfig struct {
	LogLevel          string
	LogLevelExplicit  bool
	LogFormat         string
	LogFormatExplicit bool
}

func loggingFlagConfigFromCommand(cmd *cobra.Command) loggingFlagConfig {
	if cmd == nil || cmd.Root() == nil {
		return loggingFlagConfig{}
	}
	flags := cmd.Root().PersistentFlags()
	level, _ := flags.GetString("log-level")
	format, _ := flags.GetString("log-format")
	return loggingFlagConfig{
		LogLevel:          level,
		LogLevelExplicit:  flags.Changed("log-level"),
		LogFormat:         format,
		LogFormatExplicit: flags.Changed("log-format"),
	}
}

func applyLoggingEnv(config loggingFlagConfig) error {
	if config.LogLevelExplicit {
		if err := os.Setenv("AI_SETUP_LOG_LEVEL", config.LogLevel); err != nil {
			return err
		}
	}
	if config.LogFormatExplicit {
		if err := os.Setenv("AI_SETUP_LOG_FORMAT", config.LogFormat); err != nil {
			return err
		}
	}
	orchlog.Configure("", "")
	return nil
}

func daemonLogger() *charmlog.Logger {
	return orchlog.Default().With("component", "daemon")
}

// ──────────────────── Implementation ──────────────────────────────

func runStdio(cmd *cobra.Command) error {
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}
	runtimeConfig, err := runtimeConfigForCommand(cmd)
	if err != nil {
		return err
	}

	database, err := openDatabase("")
	if err != nil {
		return err
	}
	defer database.Close()

	sc := orchmcp.NewScopeContext(scope, projectRoot, globalRoot)
	o := orchmcp.NewOrchestrator(database, sc, orchmcp.WithRuntimeConfig(runtimeConfig))

	mcpServer := server.NewMCPServer("lazyai-orchestrator", "0.1.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithLogging(),
	)
	o.RegisterTools(mcpServer)
	return server.ServeStdio(mcpServer)
}

func runServe(cmd *cobra.Command, args []string) error {
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}
	runtimeConfig, err := runtimeConfigForCommand(cmd)
	if err != nil {
		return err
	}
	resolvedIdleTimeout, err := resolveIdleTimeout(cmd)
	if err != nil {
		return err
	}

	if detachFlag {
		return startDetached(port, projectRoot, scope, execMode, configPath, cmd.Flags().Changed("execution-mode"), cmd.Flags().Changed("config"), resolvedIdleTimeout)
	}

	database, err := openDatabase(projectRoot)
	if err != nil {
		return err
	}
	defer database.Close()

	sc := orchmcp.NewScopeContext(scope, projectRoot, globalRoot)
	o := orchmcp.NewOrchestrator(database, sc, orchmcp.WithRuntimeConfig(runtimeConfig))

	mcpServer := server.NewMCPServer("lazyai-orchestrator", "0.1.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithLogging(),
	)
	o.RegisterTools(mcpServer)

	// Start background queue worker
	w := &queue.Worker{
		DB:              database,
		Queue:           o.Queue,
		PollInterval:    2 * time.Second,
		ReclaimInterval: 30 * time.Second,
		ReclaimTimeoutMs: 60000,
	}
	// Register a no-op handler that logs job types; real handlers can be registered by the MCP layer
	w.RegisterHandler("noop", &noopJobHandler{})
	workerCtx, workerStop := context.WithCancel(context.Background())
	go w.Start(workerCtx)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	mux := http.NewServeMux()
	httpSrv := &http.Server{Addr: addr, Handler: mux, ErrorLog: daemonLogger().With("component", "http").StandardLog()}
	mcpHTTPServer := server.NewStreamableHTTPServer(mcpServer, server.WithStreamableHTTPServer(httpSrv))
	startedAt := time.Now().UTC().Format(time.RFC3339)
	var shutdownOnce sync.Once
	tracker := newClientTracker(resolvedIdleTimeout)
	var idle *idleManager

	shutdown := func(reason string) {
		shutdownOnce.Do(func() {
			go func() {
				daemonLogger().Info("shutting down orchestrator daemon", "reason", reason)
				workerStop()
				clearDiscovery()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := mcpHTTPServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
					daemonLogger().Error("shutdown error", "error", err)
				}
			}()
		})
	}

	idle = newIdleManager(idleManagerOptions{
		Timeout: resolvedIdleTimeout,
		Tracker: tracker,
		ActiveRuns: func(context.Context) (db.ActiveRunCounts, error) {
			return database.ActiveRunCounts()
		},
		Shutdown: shutdown,
	})

	registerServeRoutes(mux, serveRouteConfig{
		Port:          port,
		ProjectRoot:   projectRoot,
		Scope:         scope,
		RuntimeConfig: runtimeConfig,
		StartedAt:     startedAt,
		Database:      database,
		Orchestrator:  o,
		Tracker:       tracker,
		Idle:          idle,
		MCPHandler:    mcpHTTPServer,
		Shutdown:      shutdown,
	})

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	// Write discovery file
	writeDiscovery(port, os.Getpid())
	defer clearDiscovery()

	daemonLogger().Info("orchestrator daemon listening", "url", fmt.Sprintf("http://127.0.0.1:%d/mcp", port), "port", port)

	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	idle.start(ctx)

	go func() {
		<-ctx.Done()
		shutdown("signal")
	}()

	if err := httpSrv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

type serveRouteConfig struct {
	Port          int
	ProjectRoot   string
	Scope         string
	RuntimeConfig orchmcp.RuntimeConfig
	StartedAt     string
	Database      *db.DB
	Orchestrator  *orchmcp.Orchestrator
	Tracker       *clientTracker
	Idle          *idleManager
	MCPHandler    http.Handler
	Shutdown      func(string)
}

func registerServeRoutes(mux *http.ServeMux, config serveRouteConfig) {
	mux.Handle("/mcp", config.Tracker.trackHTTP("mcp", config.MCPHandler))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(config.daemonHealth(r.Context()))
	})
	mux.HandleFunc("/admin/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "shutting_down"})
		config.Shutdown("admin request")
	})

	dashboard.RegisterViewRoutes(mux, dashboard.ViewConfig{})
	mux.Handle("/api/dashboard/", dashboard.NewHandler(dashboard.HandlerConfig{
		ReadModel: dashboard.NewReadModel(config.Database),
		Catalog:   dashboard.NewCatalogAdapter(config.Orchestrator.Catalog),
		Events:    config.Orchestrator.Events,
		Health: func(ctx context.Context) dashboard.HealthView {
			return dashboardHealthView(config.daemonHealth(ctx))
		},
	}))
}

func (config serveRouteConfig) daemonHealth(ctx context.Context) daemonHealth {
	clients := config.Tracker.snapshot()
	idleState := config.Idle.status(ctx)
	return daemonHealth{
		Status:        "ok",
		Name:          "lazyai-orchestrator",
		Port:          config.Port,
		PID:           os.Getpid(),
		StartedAt:     config.StartedAt,
		ProjectRoot:   config.ProjectRoot,
		Scope:         config.Scope,
		ExecutionMode: string(config.RuntimeConfig.ExecutionMode),
		ConfigPath:    config.RuntimeConfig.ConfigPath,
		Clients:       clients,
		Idle:          idleState,
		ActiveRuns:    idleState.ActiveRuns,
	}
}

func dashboardHealthView(health daemonHealth) dashboard.HealthView {
	return dashboard.HealthView{
		Status:        health.Status,
		Name:          health.Name,
		Port:          health.Port,
		PID:           health.PID,
		StartedAt:     health.StartedAt,
		ProjectRoot:   health.ProjectRoot,
		Scope:         health.Scope,
		ExecutionMode: health.ExecutionMode,
		ConfigPath:    health.ConfigPath,
		Clients:       health.Clients,
		Idle:          health.Idle,
		ActiveRuns:    health.ActiveRuns,
	}
}

func runConnect(cmd *cobra.Command, args []string) error {
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}
	if _, err := runtimeConfigForCommand(cmd); err != nil {
		return err
	}
	resolvedIdleTimeout, err := resolveIdleTimeout(cmd)
	if err != nil {
		return err
	}

	info, err := findRunningServer()
	if err != nil {
		// Start daemon in background
		if err := startDetachedQuiet(port, projectRoot, scope, execMode, configPath, cmd.Flags().Changed("execution-mode"), cmd.Flags().Changed("config"), resolvedIdleTimeout); err != nil {
			info, findErr := findRunningServer()
			if findErr != nil {
				return fmt.Errorf("failed to start daemon: %w", err)
			}
			return runStdioBridge(fmt.Sprintf("http://127.0.0.1:%d/mcp", info.Port))
		}
		info, err = findRunningServer()
		if err != nil {
			return fmt.Errorf("daemon did not become ready: %w", err)
		}
	}

	// Create stdio bridge to daemon
	return runStdioBridge(fmt.Sprintf("http://127.0.0.1:%d/mcp", info.Port))
}

func runStart(cmd *cobra.Command, args []string) error {
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}
	if _, err := runtimeConfigForCommand(cmd); err != nil {
		return err
	}
	resolvedIdleTimeout, err := resolveIdleTimeout(cmd)
	if err != nil {
		return err
	}
	info, err := findRunningServer()
	if err == nil {
		fmt.Printf("Orchestrator already running at http://127.0.0.1:%d/mcp (pid %d)\n", info.Port, info.PID)
		return nil
	}
	return startDetached(port, projectRoot, scope, execMode, configPath, cmd.Flags().Changed("execution-mode"), cmd.Flags().Changed("config"), resolvedIdleTimeout)
}

func runStop(cmd *cobra.Command, args []string) error {
	info, err := findRunningServer()
	if err != nil {
		return fmt.Errorf("no running orchestrator found")
	}
	fmt.Printf("Stopping orchestrator (pid %d)...\n", info.PID)
	if err := requestShutdown(info.Port); err != nil {
		return err
	}
	if !waitForServerStop(info.Port, 5*time.Second) {
		return fmt.Errorf("shutdown requested, but orchestrator on port %d is still responding", info.Port)
	}
	clearDiscovery()
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	info, err := findRunningServer()
	if err != nil {
		fmt.Println("Orchestrator daemon: not running")
		return nil
	}
	fmt.Printf("Orchestrator daemon: running (pid %d, port %d)\n", info.PID, info.Port)
	health, err := checkServerHealth(info.Port, 2*time.Second)
	if err != nil {
		return fmt.Errorf("daemon health unavailable: %w", err)
	}
	if health.ExecutionMode != "" {
		fmt.Printf("Execution mode: %s\n", health.ExecutionMode)
	}
	if health.ConfigPath != "" {
		fmt.Printf("Config: %s\n", health.ConfigPath)
	}
	fmt.Printf("Clients: %d active/recent (%d active requests, %d recent MCP sessions)\n", health.Clients.Count, health.Clients.ActiveRequests, health.Clients.RecentMCPSessions)
	fmt.Printf("Active runs/jobs: %d (chains %d, teams %d, workflows %d, queue jobs %d)\n", health.ActiveRuns.Total, health.ActiveRuns.Chains, health.ActiveRuns.Teams, health.ActiveRuns.Workflows, health.ActiveRuns.QueueJobs)
	if health.Idle.Enabled {
		fmt.Printf("Idle shutdown: enabled (timeout %ds, idle for %ds, shutdown after %ds)\n", health.Idle.TimeoutSeconds, health.Idle.IdleForSeconds, health.Idle.ShutdownAfterSeconds)
	} else {
		fmt.Println("Idle shutdown: disabled")
	}
	if len(health.Idle.BlockingReasons) > 0 {
		fmt.Printf("Idle blockers: %s\n", strings.Join(health.Idle.BlockingReasons, ", "))
	}
	return nil
}

// ──────────────────── helpers ─────────────────────────────────────

func openDatabase(projectRoot string) (*db.DB, error) {
	path := ":memory:"
	if dbPath := os.Getenv("AI_SETUP_ORCHESTRATOR_DB"); dbPath != "" {
		path = dbPath
	} else if projectRoot != "" {
		if err := os.MkdirAll(dataDir(), 0755); err != nil {
			return nil, err
		}
		path = filepath.Join(dataDir(), "orchestrator.db")
	}

	database, err := db.Open(path)
	if err != nil {
		return nil, err
	}
	if err := database.RunMigrations(); err != nil {
		return nil, err
	}
	return database, nil
}

func resolveIdleTimeout(cmd *cobra.Command) (time.Duration, error) {
	if cmd == nil || cmd.Flags().Changed("idle-timeout") {
		return idleTimeout, nil
	}
	value := strings.TrimSpace(os.Getenv("AI_SETUP_ORCHESTRATOR_IDLE_TIMEOUT"))
	if value == "" {
		return idleTimeout, nil
	}
	parsed, err := parseDurationSetting(value)
	if err != nil {
		return 0, fmt.Errorf("invalid AI_SETUP_ORCHESTRATOR_IDLE_TIMEOUT %q: %w", value, err)
	}
	return parsed, nil
}

func parseDurationSetting(value string) (time.Duration, error) {
	if value == "0" {
		return 0, nil
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0, fmt.Errorf("must be non-negative")
		}
		return time.Duration(seconds) * time.Second, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, fmt.Errorf("must be non-negative")
	}
	return parsed, nil
}

func runtimeConfigForCommand(cmd *cobra.Command) (orchmcp.RuntimeConfig, error) {
	modeExplicit := false
	configExplicit := false
	if cmd != nil {
		modeExplicit = cmd.Flags().Changed("execution-mode")
		configExplicit = cmd.Flags().Changed("config")
	}
	return orchmcp.LoadRuntimeConfig(orchmcp.RuntimeConfigOptions{
		ProjectRoot:           projectRoot,
		ConfigPath:            configPath,
		ConfigPathExplicit:    configExplicit,
		ExecutionMode:         execMode,
		ExecutionModeExplicit: modeExplicit,
	})
}

func startDetached(port int, projectRoot, scope, execMode, configPath string, execModeExplicit, configPathExplicit bool, idleTimeout time.Duration) error {
	return startDetachedWithStatus(port, projectRoot, scope, execMode, configPath, execModeExplicit, configPathExplicit, idleTimeout, os.Stdout)
}

func startDetachedQuiet(port int, projectRoot, scope, execMode, configPath string, execModeExplicit, configPathExplicit bool, idleTimeout time.Duration) error {
	return startDetachedWithStatus(port, projectRoot, scope, execMode, configPath, execModeExplicit, configPathExplicit, idleTimeout, nil)
}

func startDetachedWithStatus(port int, projectRoot, scope, execMode, configPath string, execModeExplicit, configPathExplicit bool, idleTimeout time.Duration, status io.Writer) error {
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}
	if _, err := orchmcp.LoadRuntimeConfig(orchmcp.RuntimeConfigOptions{
		ProjectRoot:           projectRoot,
		ConfigPath:            configPath,
		ConfigPathExplicit:    configPathExplicit,
		ExecutionMode:         execMode,
		ExecutionModeExplicit: execModeExplicit,
	}); err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	if err := os.MkdirAll(dataDir(), 0755); err != nil {
		return err
	}
	logPath := filepath.Join(dataDir(), "daemon.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open daemon log: %w", err)
	}
	defer logFile.Close()

	args := []string{
		"serve",
		"--port", strconv.Itoa(port),
		"--project", projectRoot,
		"--scope", scope,
		"--idle-timeout", idleTimeout.String(),
	}
	if execModeExplicit {
		args = append(args, "--execution-mode", execMode)
	}
	if configPathExplicit && strings.TrimSpace(configPath) != "" {
		args = append(args, "--config", configPath)
	}
	cmd := exec.Command(exe, args...)
	cmd.Env = os.Environ()
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	configureDetachedCommand(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("release daemon process: %w", err)
	}

	if err := waitForServerHealth(port, pid, 5*time.Second); err != nil {
		if processExists(pid) {
			terminateProcess(pid)
		}
		return fmt.Errorf("orchestrator did not become ready on port %d; see %s: %w", port, logPath, err)
	}

	if status != nil {
		fmt.Fprintf(status, "Orchestrator started at http://127.0.0.1:%d/mcp (logs: %s)\n", port, logPath)
	}
	return nil
}

// ──────────────────── discovery ────────────────────────────────────

type daemonInfo struct {
	Port      int    `json:"port"`
	PID       int    `json:"pid"`
	StartedAt string `json:"startedAt"`
}

type daemonHealth struct {
	Status        string             `json:"status"`
	Name          string             `json:"name"`
	Port          int                `json:"port"`
	PID           int                `json:"pid"`
	StartedAt     string             `json:"startedAt"`
	ProjectRoot   string             `json:"projectRoot,omitempty"`
	Scope         string             `json:"scope,omitempty"`
	ExecutionMode string             `json:"executionMode,omitempty"`
	ConfigPath    string             `json:"configPath,omitempty"`
	Clients       clientSnapshot     `json:"clients"`
	Idle          idleStatus         `json:"idle"`
	ActiveRuns    db.ActiveRunCounts `json:"activeRuns"`
}

func dataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "lazyai-orchestrator")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.TempDir()
	}
	return filepath.Join(home, ".local", "share", "lazyai-orchestrator")
}

func discoveryPath() string {
	return filepath.Join(dataDir(), "daemon.json")
}

func writeDiscovery(port, pid int) {
	info := daemonInfo{Port: port, PID: pid, StartedAt: time.Now().UTC().Format(time.RFC3339)}
	b, _ := json.MarshalIndent(info, "", "  ")
	os.MkdirAll(filepath.Dir(discoveryPath()), 0755)
	os.WriteFile(discoveryPath(), b, 0644)
}

func findRunningServer() (*daemonInfo, error) {
	b, err := os.ReadFile(discoveryPath())
	if err != nil {
		return nil, err
	}
	var info daemonInfo
	if err := json.Unmarshal(b, &info); err != nil {
		return nil, err
	}
	if info.Port <= 0 || info.PID <= 0 {
		clearDiscovery()
		return nil, fmt.Errorf("stale discovery: invalid daemon metadata")
	}
	if !processExists(info.PID) {
		clearDiscovery()
		return nil, fmt.Errorf("stale discovery: process %d is not running", info.PID)
	}
	health, err := checkServerHealth(info.Port, 750*time.Millisecond)
	if err != nil {
		if isDefinitiveHealthFailure(err) {
			clearDiscovery()
		}
		return nil, fmt.Errorf("stale discovery: health check failed: %w", err)
	}
	if health.PID != 0 && health.PID != info.PID {
		clearDiscovery()
		return nil, fmt.Errorf("stale discovery: health pid %d does not match discovery pid %d", health.PID, info.PID)
	}
	return &info, nil
}

func isDefinitiveHealthFailure(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET)
}

func clearDiscovery() {
	os.Remove(discoveryPath())
}

func checkServerHealth(port int, timeout time.Duration) (*daemonHealth, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected health status %d", resp.StatusCode)
	}
	var health daemonHealth
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, err
	}
	if health.Status != "ok" || health.Name != "lazyai-orchestrator" {
		return nil, fmt.Errorf("unexpected health response")
	}
	return &health, nil
}

func waitForServerHealth(port, expectedPID int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if health, err := checkServerHealth(port, 250*time.Millisecond); err == nil {
			if expectedPID == 0 || health.PID == expectedPID {
				return nil
			}
			lastErr = fmt.Errorf("health pid %d does not match daemon pid %d", health.PID, expectedPID)
		} else {
			lastErr = err
		}
		time.Sleep(150 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("timed out")
	}
	return lastErr
}

func waitForServerStop(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := checkServerHealth(port, 200*time.Millisecond); err != nil {
			return true
		}
		time.Sleep(150 * time.Millisecond)
	}
	return false
}

func requestShutdown(port int) error {
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/admin/shutdown", port), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("admin shutdown request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("admin shutdown returned status %d", resp.StatusCode)
	}
	return nil
}

// ────────────────────── misc ─────────────────────────────────────

// ServeStdio is a helper for stdio transport.
func init() {
	// Ensure server.ServeStdio exists by using the mcp-go API
	_ = server.ServeStdio
}

// noopJobHandler is a placeholder handler that logs and immediately succeeds.
// Real job handlers are registered by the MCP layer or future workers.
type noopJobHandler struct{}

func (h *noopJobHandler) Handle(ctx context.Context, job *queue.Job) error {
	daemonLogger().Info("noop job handled", "jobId", job.ID, "jobType", job.JobType)
	return nil
}
