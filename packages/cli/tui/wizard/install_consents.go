package wizard

import (
	"fmt"
	"sort"
	"strings"
)

type installConsentHint struct {
	Server  string
	Command string
	Hint    string
}

func missingInstallConsents(selectedServers []string) []installConsentHint {
	catalog, err := loadMcpCatalog()
	if err != nil || len(selectedServers) == 0 {
		return nil
	}

	selected := make(map[string]struct{}, len(selectedServers))
	for _, id := range selectedServers {
		if id == "" {
			continue
		}
		selected[id] = struct{}{}
	}
	if len(selected) == 0 {
		return nil
	}

	serverIDs := make([]string, 0, len(selected))
	for id := range selected {
		serverIDs = append(serverIDs, id)
	}
	sort.Strings(serverIDs)

	consents := make([]installConsentHint, 0)
	for _, id := range serverIDs {
		server, ok := catalog.Servers[id]
		if !ok || !server.RequiresInstall {
			continue
		}

		command := strings.TrimSpace(server.CliEquivalent)
		if command == "" {
			command = id
		}

		detectedPath, err := cliToolLookPath(command)
		if err == nil && detectedPath != "" {
			continue
		}

		hint := strings.TrimSpace(server.InstallHint)
		if hint == "" {
			if tool, ok := catalog.CliTools[id]; ok {
				hint = strings.TrimSpace(tool.InstallHint)
			}
		}
		if hint == "" {
			continue
		}

		consents = append(consents, installConsentHint{
			Server:  id,
			Command: command,
			Hint:    hint,
		})
	}

	return consents
}

func formatInstallConsents(consents []installConsentHint) []string {
	if len(consents) == 0 {
		return nil
	}

	lines := make([]string, 0, len(consents))
	for _, consent := range consents {
		if consent.Server == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s", consent.Server, consent.Hint))
	}
	return lines
}
