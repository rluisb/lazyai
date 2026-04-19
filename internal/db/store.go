package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// Store provides CRUD operations for ai-setup data.
type Store struct {
	db *DB
}

// NewStore creates a new Store backed by db.
func NewStore(db *DB) *Store {
	return &Store{db: db}
}

// ReadStoreData reads all store data from SQLite and assembles a full StoreData struct.
func (s *Store) ReadStoreData() (*types.StoreData, error) {
	meta, err := s.ReadMeta()
	if err != nil {
		return nil, fmt.Errorf("read meta: %w", err)
	}

	config, err := s.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	selections, err := s.ReadSelections()
	if err != nil {
		return nil, fmt.Errorf("read selections: %w", err)
	}

	files, err := s.ReadTrackedFiles()
	if err != nil {
		return nil, fmt.Errorf("read tracked files: %w", err)
	}

	sync, err := s.ReadSync()
	if err != nil {
		return nil, fmt.Errorf("read sync: %w", err)
	}

	operations, err := s.ReadOperations()
	if err != nil {
		return nil, fmt.Errorf("read operations: %w", err)
	}

	return &types.StoreData{
		Meta:       *meta,
		Config:     *config,
		Selections: *selections,
		Files:      files,
		Sync:       *sync,
		Operations: operations,
	}, nil
}

// WriteStoreData writes all store data to SQLite within a single transaction.
func (s *Store) WriteStoreData(data *types.StoreData) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := writeMeta(tx, &data.Meta); err != nil {
		return fmt.Errorf("write meta: %w", err)
	}
	if err := writeConfig(tx, &data.Config); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := writeSelections(tx, &data.Selections); err != nil {
		return fmt.Errorf("write selections: %w", err)
	}
	if err := writeSync(tx, &data.Sync); err != nil {
		return fmt.Errorf("write sync: %w", err)
	}

	// Tracked files: clear and rewrite.
	if _, err := tx.Exec("DELETE FROM tracked_files"); err != nil {
		return fmt.Errorf("clear tracked files: %w", err)
	}
	for _, f := range data.Files {
		if err := upsertTrackedFile(tx, f); err != nil {
			return fmt.Errorf("write tracked file %s: %w", f.Path, err)
		}
	}

	// Operations: clear and rewrite.
	if _, err := tx.Exec("DELETE FROM operations"); err != nil {
		return fmt.Errorf("clear operations: %w", err)
	}
	if err := writeOperations(tx, data.Operations); err != nil {
		return fmt.Errorf("write operations: %w", err)
	}

	return tx.Commit()
}

// ---------------------------------------------------------------------------
// Meta
// ---------------------------------------------------------------------------

// ReadMeta reads the meta table.
func (s *Store) ReadMeta() (*types.Meta, error) {
	var m types.Meta
	var schemaVersion int
	var cliVersion, installedAt, lastUpdatedAt string

	err := s.db.QueryRow(
		"SELECT schema_version, cli_version, installed_at, last_updated_at FROM meta WHERE id = 1",
	).Scan(&schemaVersion, &cliVersion, &installedAt, &lastUpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("query meta: %w", err)
	}

	m.SchemaVersion = schemaVersion
	m.CLIVersion = cliVersion
	m.InstalledAt = installedAt
	m.LastUpdatedAt = lastUpdatedAt
	return &m, nil
}

func writeMeta(exec sqlExecutor, m *types.Meta) error {
	_, err := exec.Exec(`
		INSERT OR REPLACE INTO meta (id, schema_version, cli_version, installed_at, last_updated_at)
		VALUES (1, ?, ?, ?, ?)`,
		m.SchemaVersion, m.CLIVersion, m.InstalledAt, m.LastUpdatedAt)
	return err
}

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

// ReadConfig reads the config table.
func (s *Store) ReadConfig() (*types.Config, error) {
	var scope, toolsJSON, cliToolsJSON, enableServersJSON string
	var projectName, workspaceName, targetDir string
	var planningDir, planningRepoPath, reposJSON, globalRef string

	err := s.db.QueryRow(`
		SELECT scope, tools, cli_tools, enable_servers, project_name, workspace_name,
		       target_dir, planning_dir, planning_repo_path, repos, global_ref
		FROM config WHERE id = 1`,
	).Scan(&scope, &toolsJSON, &cliToolsJSON, &enableServersJSON,
		&projectName, &workspaceName, &targetDir,
		&planningDir, &planningRepoPath, &reposJSON, &globalRef)
	if err != nil {
		return nil, fmt.Errorf("query config: %w", err)
	}

	var tools []types.ToolId
	if err := json.Unmarshal([]byte(toolsJSON), &tools); err != nil {
		return nil, fmt.Errorf("unmarshal tools: %w", err)
	}

	var cliTools []string
	if err := json.Unmarshal([]byte(cliToolsJSON), &cliTools); err != nil {
		return nil, fmt.Errorf("unmarshal cliTools: %w", err)
	}

	var enableServers []string
	if err := json.Unmarshal([]byte(enableServersJSON), &enableServers); err != nil {
		return nil, fmt.Errorf("unmarshal enableServers: %w", err)
	}

	var repos []types.RepoInfo
	if err := json.Unmarshal([]byte(reposJSON), &repos); err != nil {
		return nil, fmt.Errorf("unmarshal repos: %w", err)
	}

	return &types.Config{
		SetupScope:       types.SetupScope(scope),
		Tools:            tools,
		CLITools:         cliTools,
		EnableServers:    enableServers,
		ProjectName:      projectName,
		WorkspaceName:    workspaceName,
		TargetDir:        targetDir,
		PlanningDir:      planningDir,
		PlanningRepoPath: planningRepoPath,
		Repos:            repos,
		GlobalRef:        globalRef,
	}, nil
}

func writeConfig(exec sqlExecutor, c *types.Config) error {
	toolsJSON, err := json.Marshal(c.Tools)
	if err != nil {
		return fmt.Errorf("marshal tools: %w", err)
	}
	cliToolsJSON, err := json.Marshal(c.CLITools)
	if err != nil {
		return fmt.Errorf("marshal cliTools: %w", err)
	}
	enableServersJSON, err := json.Marshal(c.EnableServers)
	if err != nil {
		return fmt.Errorf("marshal enableServers: %w", err)
	}
	reposJSON, err := json.Marshal(c.Repos)
	if err != nil {
		return fmt.Errorf("marshal repos: %w", err)
	}

	_, err = exec.Exec(`
		INSERT OR REPLACE INTO config
			(id, scope, tools, cli_tools, enable_servers, project_name, workspace_name,
			 target_dir, planning_dir, planning_repo_path, repos, global_ref)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.SetupScope, string(toolsJSON), string(cliToolsJSON), string(enableServersJSON),
		c.ProjectName, c.WorkspaceName, c.TargetDir,
		c.PlanningDir, c.PlanningRepoPath, string(reposJSON), c.GlobalRef)
	return err
}

// ---------------------------------------------------------------------------
// Selections
// ---------------------------------------------------------------------------

// ReadSelections reads the selections table.
func (s *Store) ReadSelections() (*types.WizardSelections, error) {
	var templatesJSON, rulesJSON, agentsJSON, skillsJSON, promptsJSON string
	var commandsJSON, chatmodesJSON string
	var opencodeCommandsJSON, opencodeModesJSON string
	var infraJSON, constitutionJSON, featuresJSON, gitConventionsJSON, preset string

	err := s.db.QueryRow(`
		SELECT templates, rules, agents, skills, prompts, infra, constitution,
		       features, git_conventions, preset, commands, chatmodes,
		       opencode_commands, opencode_modes
		FROM selections WHERE id = 1`,
	).Scan(&templatesJSON, &rulesJSON, &agentsJSON, &skillsJSON, &promptsJSON,
		&infraJSON, &constitutionJSON, &featuresJSON, &gitConventionsJSON, &preset,
		&commandsJSON, &chatmodesJSON,
		&opencodeCommandsJSON, &opencodeModesJSON)
	if err != nil {
		return nil, fmt.Errorf("query selections: %w", err)
	}

	var templates []types.TemplateId
	if err := json.Unmarshal([]byte(templatesJSON), &templates); err != nil {
		return nil, fmt.Errorf("unmarshal templates: %w", err)
	}
	var rules []types.RuleId
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return nil, fmt.Errorf("unmarshal rules: %w", err)
	}
	var agents []types.AgentId
	if err := json.Unmarshal([]byte(agentsJSON), &agents); err != nil {
		return nil, fmt.Errorf("unmarshal agents: %w", err)
	}
	var skills []types.SkillId
	if err := json.Unmarshal([]byte(skillsJSON), &skills); err != nil {
		return nil, fmt.Errorf("unmarshal skills: %w", err)
	}
	var prompts []types.PromptId
	if err := json.Unmarshal([]byte(promptsJSON), &prompts); err != nil {
		return nil, fmt.Errorf("unmarshal prompts: %w", err)
	}
	var infra []types.InfraId
	if err := json.Unmarshal([]byte(infraJSON), &infra); err != nil {
		return nil, fmt.Errorf("unmarshal infra: %w", err)
	}
	var constitution []string
	if err := json.Unmarshal([]byte(constitutionJSON), &constitution); err != nil {
		return nil, fmt.Errorf("unmarshal constitution: %w", err)
	}

	var features types.FeatureFlags
	if err := json.Unmarshal([]byte(featuresJSON), &features); err != nil {
		return nil, fmt.Errorf("unmarshal features: %w", err)
	}

	var gitConventions types.GitConventions
	if err := json.Unmarshal([]byte(gitConventionsJSON), &gitConventions); err != nil {
		return nil, fmt.Errorf("unmarshal gitConventions: %w", err)
	}

	var commands []types.CommandId
	if err := json.Unmarshal([]byte(commandsJSON), &commands); err != nil {
		return nil, fmt.Errorf("unmarshal commands: %w", err)
	}
	var chatmodes []types.ChatModeId
	if err := json.Unmarshal([]byte(chatmodesJSON), &chatmodes); err != nil {
		return nil, fmt.Errorf("unmarshal chatmodes: %w", err)
	}
	var opencodeCommands []types.OpenCodeCommandId
	if err := json.Unmarshal([]byte(opencodeCommandsJSON), &opencodeCommands); err != nil {
		return nil, fmt.Errorf("unmarshal opencode_commands: %w", err)
	}
	var opencodeModes []types.OpenCodeModeId
	if err := json.Unmarshal([]byte(opencodeModesJSON), &opencodeModes); err != nil {
		return nil, fmt.Errorf("unmarshal opencode_modes: %w", err)
	}

	return &types.WizardSelections{
		Templates:        templates,
		Rules:            rules,
		Agents:           agents,
		Skills:           skills,
		Prompts:          prompts,
		Commands:         commands,
		ChatModes:        chatmodes,
		OpenCodeCommands: opencodeCommands,
		OpenCodeModes:    opencodeModes,
		Infra:            infra,
		Constitution:     constitution,
		Features:         &features,
		GitConventions:   &gitConventions,
	}, nil
}

func writeSelections(exec sqlExecutor, s *types.WizardSelections) error {
	templatesJSON, err := json.Marshal(s.Templates)
	if err != nil {
		return fmt.Errorf("marshal templates: %w", err)
	}
	rulesJSON, err := json.Marshal(s.Rules)
	if err != nil {
		return fmt.Errorf("marshal rules: %w", err)
	}
	agentsJSON, err := json.Marshal(s.Agents)
	if err != nil {
		return fmt.Errorf("marshal agents: %w", err)
	}
	skillsJSON, err := json.Marshal(s.Skills)
	if err != nil {
		return fmt.Errorf("marshal skills: %w", err)
	}
	promptsJSON, err := json.Marshal(s.Prompts)
	if err != nil {
		return fmt.Errorf("marshal prompts: %w", err)
	}
	infraJSON, err := json.Marshal(s.Infra)
	if err != nil {
		return fmt.Errorf("marshal infra: %w", err)
	}
	constitutionJSON, err := json.Marshal(s.Constitution)
	if err != nil {
		return fmt.Errorf("marshal constitution: %w", err)
	}

	var featuresJSON []byte
	if s.Features != nil {
		featuresJSON, err = json.Marshal(s.Features)
		if err != nil {
			return fmt.Errorf("marshal features: %w", err)
		}
	} else {
		featuresJSON = []byte("{}")
	}

	var gitConventionsJSON []byte
	if s.GitConventions != nil {
		gitConventionsJSON, err = json.Marshal(s.GitConventions)
		if err != nil {
			return fmt.Errorf("marshal gitConventions: %w", err)
		}
	} else {
		gitConventionsJSON = []byte("{}")
	}

	commandsJSON, err := json.Marshal(s.Commands)
	if err != nil {
		return fmt.Errorf("marshal commands: %w", err)
	}
	chatmodesJSON, err := json.Marshal(s.ChatModes)
	if err != nil {
		return fmt.Errorf("marshal chatmodes: %w", err)
	}
	opencodeCommandsJSON, err := json.Marshal(s.OpenCodeCommands)
	if err != nil {
		return fmt.Errorf("marshal opencode_commands: %w", err)
	}
	opencodeModesJSON, err := json.Marshal(s.OpenCodeModes)
	if err != nil {
		return fmt.Errorf("marshal opencode_modes: %w", err)
	}

	preset := "standard"

	_, err = exec.Exec(`
		INSERT OR REPLACE INTO selections
			(id, templates, rules, agents, skills, prompts, infra, constitution,
			 features, git_conventions, preset, commands, chatmodes,
			 opencode_commands, opencode_modes)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(templatesJSON), string(rulesJSON), string(agentsJSON), string(skillsJSON),
		string(promptsJSON), string(infraJSON), string(constitutionJSON),
		string(featuresJSON), string(gitConventionsJSON), preset,
		string(commandsJSON), string(chatmodesJSON),
		string(opencodeCommandsJSON), string(opencodeModesJSON))
	return err
}

// ---------------------------------------------------------------------------
// TrackedFiles
// ---------------------------------------------------------------------------

// ReadTrackedFiles reads all tracked files.
func (s *Store) ReadTrackedFiles() ([]types.TrackedFile, error) {
	rows, err := s.db.Query(`
		SELECT path, hash, source, owner, status, installed_at, last_checked_at
		FROM tracked_files ORDER BY path`)
	if err != nil {
		return nil, fmt.Errorf("query tracked files: %w", err)
	}
	defer rows.Close()

	var files []types.TrackedFile
	for rows.Next() {
		var f types.TrackedFile
		var owner, status string
		if err := rows.Scan(&f.Path, &f.Hash, &f.Source, &owner, &status,
			&f.InstalledAt, &f.LastCheckedAt); err != nil {
			return nil, fmt.Errorf("scan tracked file: %w", err)
		}
		f.Owner = types.FileOwner(owner)
		f.Status = types.FileStatus(status)
		files = append(files, f)
	}

	if files == nil {
		files = []types.TrackedFile{}
	}
	return files, rows.Err()
}

// UpsertTrackedFile inserts or updates a tracked file.
func (s *Store) UpsertTrackedFile(file types.TrackedFile) error {
	return upsertTrackedFile(s.db, file)
}

func upsertTrackedFile(db sqlExecutor, file types.TrackedFile) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO tracked_files
			(path, hash, source, owner, status, installed_at, last_checked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		file.Path, file.Hash, file.Source, string(file.Owner), string(file.Status),
		file.InstalledAt, file.LastCheckedAt)
	return err
}

// DeleteTrackedFile removes a tracked file by path.
func (s *Store) DeleteTrackedFile(path string) error {
	_, err := s.db.Exec("DELETE FROM tracked_files WHERE path = ?", path)
	return err
}

// ---------------------------------------------------------------------------
// Operations
// ---------------------------------------------------------------------------

// ReadOperations reads all operations.
func (s *Store) ReadOperations() ([]types.Operation, error) {
	rows, err := s.db.Query(`
		SELECT id, type, timestamp, files_affected, result, backup_paths, error
		FROM operations ORDER BY timestamp DESC`)
	if err != nil {
		return nil, fmt.Errorf("query operations: %w", err)
	}
	defer rows.Close()

	var operations []types.Operation
	for rows.Next() {
		var op types.Operation
		var filesAffectedJSON, backupPathsJSON string
		var resultStr string
		var errorStr sql.NullString

		if err := rows.Scan(&op.ID, &op.Type, &op.Timestamp,
			&filesAffectedJSON, &resultStr, &backupPathsJSON, &errorStr); err != nil {
			return nil, fmt.Errorf("scan operation: %w", err)
		}

		if err := json.Unmarshal([]byte(filesAffectedJSON), &op.FilesAffected); err != nil {
			return nil, fmt.Errorf("unmarshal filesAffected: %w", err)
		}
		op.Result = types.OperationResult(resultStr)
		if err := json.Unmarshal([]byte(backupPathsJSON), &op.BackupPaths); err != nil {
			return nil, fmt.Errorf("unmarshal backupPaths: %w", err)
		}
		if errorStr.Valid {
			op.Error = errorStr.String
		}

		operations = append(operations, op)
	}

	if operations == nil {
		operations = []types.Operation{}
	}
	return operations, rows.Err()
}

// AppendOperation adds an operation. It caps the total at 50 by deleting the
// oldest entries when over the limit.
func (s *Store) AppendOperation(op types.Operation) error {
	filesAffectedJSON, err := json.Marshal(op.FilesAffected)
	if err != nil {
		return fmt.Errorf("marshal filesAffected: %w", err)
	}
	backupPathsJSON, err := json.Marshal(op.BackupPaths)
	if err != nil {
		return fmt.Errorf("marshal backupPaths: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO operations (id, type, timestamp, files_affected, result, backup_paths, error)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		op.ID, op.Type, op.Timestamp, string(filesAffectedJSON),
		string(op.Result), string(backupPathsJSON), op.Error)
	if err != nil {
		return fmt.Errorf("insert operation: %w", err)
	}

	// Cap at 50: delete oldest if over limit.
	_, err = tx.Exec(`
		DELETE FROM operations WHERE id IN (
			SELECT id FROM operations ORDER BY timestamp DESC LIMIT -1 OFFSET 50
		)`)
	if err != nil {
		return fmt.Errorf("trim operations: %w", err)
	}

	return tx.Commit()
}

func writeOperations(tx *sql.Tx, ops []types.Operation) error {
	for _, op := range ops {
		filesAffectedJSON, err := json.Marshal(op.FilesAffected)
		if err != nil {
			return fmt.Errorf("marshal filesAffected: %w", err)
		}
		backupPathsJSON, err := json.Marshal(op.BackupPaths)
		if err != nil {
			return fmt.Errorf("marshal backupPaths: %w", err)
		}

		_, err = tx.Exec(`
			INSERT INTO operations (id, type, timestamp, files_affected, result, backup_paths, error)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			op.ID, op.Type, op.Timestamp, string(filesAffectedJSON),
			string(op.Result), string(backupPathsJSON), op.Error)
		if err != nil {
			return fmt.Errorf("insert operation %s: %w", op.ID, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Sync
// ---------------------------------------------------------------------------

// ReadSync reads the sync state.
func (s *Store) ReadSync() (*types.Sync, error) {
	var lastSyncAt string
	var dirty int

	err := s.db.QueryRow(
		"SELECT last_sync_at, dirty FROM sync WHERE id = 1",
	).Scan(&lastSyncAt, &dirty)
	if err != nil {
		return nil, fmt.Errorf("query sync: %w", err)
	}

	return &types.Sync{
		LastSyncAt: lastSyncAt,
		Dirty:      dirty != 0,
	}, nil
}

func writeSync(tx *sql.Tx, s *types.Sync) error {
	dirty := 0
	if s.Dirty {
		dirty = 1
	}
	_, err := tx.Exec(`
		INSERT OR REPLACE INTO sync (id, last_sync_at, dirty) VALUES (1, ?, ?)`,
		s.LastSyncAt, dirty)
	return err
}

// ---------------------------------------------------------------------------
// FeatureFlags
// ---------------------------------------------------------------------------

// ReadFeatureFlag reads a single feature flag by key.
func (s *Store) ReadFeatureFlag(key string) (bool, error) {
	var value int
	err := s.db.QueryRow(
		"SELECT value FROM feature_flags WHERE key = ?", key,
	).Scan(&value)
	if err != nil {
		return false, fmt.Errorf("query feature flag %s: %w", key, err)
	}
	return value != 0, nil
}

// SetFeatureFlag sets a feature flag.
func (s *Store) SetFeatureFlag(key string, value bool) error {
	v := 0
	if value {
		v = 1
	}
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO feature_flags (key, value) VALUES (?, ?)", key, v)
	return err
}

// ---------------------------------------------------------------------------
// Initialize — writes default rows if tables are empty
// ---------------------------------------------------------------------------

// Initialize writes default rows for meta, config, selections, and sync
// if they don't already exist. This should be called after migrations.
func (s *Store) Initialize(cliVersion string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Meta
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM meta").Scan(&count); err != nil {
		return fmt.Errorf("count meta: %w", err)
	}
	if count == 0 {
		_, err := s.db.Exec(`
			INSERT INTO meta (id, schema_version, cli_version, installed_at, last_updated_at)
			VALUES (1, ?, ?, ?, ?)`,
			types.CurrentSchemaVersion, cliVersion, now, now)
		if err != nil {
			return fmt.Errorf("init meta: %w", err)
		}
	}

	// Config
	if err := s.db.QueryRow("SELECT COUNT(*) FROM config").Scan(&count); err != nil {
		return fmt.Errorf("count config: %w", err)
	}
	if count == 0 {
		defaults := types.DefaultStoreData()
		if err := writeConfig(s.db, &defaults.Config); err != nil {
			return fmt.Errorf("init config: %w", err)
		}
	}

	// Selections
	if err := s.db.QueryRow("SELECT COUNT(*) FROM selections").Scan(&count); err != nil {
		return fmt.Errorf("count selections: %w", err)
	}
	if count == 0 {
		defaults := types.DefaultStoreData()
		if err := writeSelections(s.db, &defaults.Selections); err != nil {
			return fmt.Errorf("init selections: %w", err)
		}
	}

	// Sync
	if err := s.db.QueryRow("SELECT COUNT(*) FROM sync").Scan(&count); err != nil {
		return fmt.Errorf("count sync: %w", err)
	}
	if count == 0 {
		_, err := s.db.Exec(`
			INSERT INTO sync (id, last_sync_at, dirty) VALUES (1, ?, 1)`, now)
		if err != nil {
			return fmt.Errorf("init sync: %w", err)
		}
	}

	return nil
}

// sqlExecutor is the common interface between *sql.DB and *sql.Tx.
type sqlExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}
