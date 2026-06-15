package cmd

import "github.com/spf13/cobra"

type commandCategory string

const (
	commandCategorySetupCore       commandCategory = "setup-core"
	commandCategoryOpsRuntimeExtra commandCategory = "ops-runtime-extra"
	commandCategoryDevHarness      commandCategory = "dev-harness"
	commandCategoryRetiredArchived commandCategory = "retired/archived"
)

type rootCommandCategoryEntry struct {
	name     string
	category commandCategory
}

var rootCommandCategoryEntries = [...]rootCommandCategoryEntry{
	{name: "add", category: commandCategorySetupCore},
	{name: "auth", category: commandCategoryOpsRuntimeExtra},
	{name: "backup", category: commandCategoryOpsRuntimeExtra},
	{name: "build-plugin", category: commandCategorySetupCore},
	{name: "compile", category: commandCategorySetupCore},
	{name: "completion", category: commandCategorySetupCore},
	{name: "completions", category: commandCategoryRetiredArchived},
	{name: "config", category: commandCategorySetupCore},
	{name: "cost", category: commandCategoryOpsRuntimeExtra},
	{name: "create", category: commandCategorySetupCore},
	{name: "doctor", category: commandCategorySetupCore},
	{name: "eject", category: commandCategorySetupCore},
	{name: "git", category: commandCategoryOpsRuntimeExtra},
	{name: "import", category: commandCategorySetupCore},
	{name: "info", category: commandCategorySetupCore},
	{name: "init", category: commandCategorySetupCore},
	{name: "ledger", category: commandCategoryOpsRuntimeExtra},
	{name: "list", category: commandCategorySetupCore},
	{name: "memory", category: commandCategoryOpsRuntimeExtra},
	{name: "message", category: commandCategoryOpsRuntimeExtra},
	{name: "metrics", category: commandCategoryOpsRuntimeExtra},
	{name: "migrate", category: commandCategorySetupCore},
	{name: "models", category: commandCategoryDevHarness},
	{name: "notify", category: commandCategoryOpsRuntimeExtra},
	{name: "restore-runtime-db", category: commandCategoryOpsRuntimeExtra},
	{name: "secret", category: commandCategoryOpsRuntimeExtra},
	{name: "server", category: commandCategorySetupCore},
	{name: "session", category: commandCategoryOpsRuntimeExtra},
	{name: "setup", category: commandCategorySetupCore},
	{name: "sidecar", category: commandCategorySetupCore},
	{name: "status", category: commandCategorySetupCore},
	{name: "update", category: commandCategorySetupCore},
	{name: "update-self", category: commandCategorySetupCore},
	{name: "validate", category: commandCategorySetupCore},
	{name: "workspace", category: commandCategorySetupCore},
}

func rootCommandCategory(cmd *cobra.Command) (commandCategory, bool) {
	if cmd == nil {
		return "", false
	}
	name := cmd.Name()
	for i := range rootCommandCategoryEntries {
		if rootCommandCategoryEntries[i].name == name {
			return rootCommandCategoryEntries[i].category, true
		}
	}
	return "", false
}

func isKnownCommandCategory(category commandCategory) bool {
	switch category {
	case commandCategorySetupCore,
		commandCategoryOpsRuntimeExtra,
		commandCategoryDevHarness,
		commandCategoryRetiredArchived:
		return true
	default:
		return false
	}
}
