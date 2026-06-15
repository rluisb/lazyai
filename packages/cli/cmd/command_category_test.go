package cmd

import "testing"

func TestRootCommandCategories(t *testing.T) {
	registered := make(map[string]struct{}, len(rootCmd.Commands()))

	for _, command := range rootCmd.Commands() {
		name := command.Name()
		registered[name] = struct{}{}

		category, ok := rootCommandCategory(command)
		if !ok {
			t.Errorf("root command %q is missing a command category", name)
			continue
		}
		if !isKnownCommandCategory(category) {
			t.Errorf("root command %q uses unknown command category %q", name, category)
		}
		if category == commandCategoryRetiredArchived && !command.Hidden {
			t.Errorf("root command %q uses %q but is not hidden", name, category)
		}
	}

	seenEntries := make(map[string]struct{}, len(rootCommandCategoryEntries))
	for i := range rootCommandCategoryEntries {
		entry := rootCommandCategoryEntries[i]
		if _, duplicate := seenEntries[entry.name]; duplicate {
			t.Errorf("root command category table has duplicate entry for %q", entry.name)
		}
		seenEntries[entry.name] = struct{}{}
		if !isKnownCommandCategory(entry.category) {
			t.Errorf("root command category table maps %q to unknown category %q", entry.name, entry.category)
		}
		if _, ok := registered[entry.name]; !ok {
			t.Errorf("root command category table includes unregistered command %q", entry.name)
		}
	}
}
