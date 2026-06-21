package wizard

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMissingInstallConsentsUsesCatalogInstallHints(t *testing.T) {
	originalLookPath := cliToolLookPath
	defer func() { cliToolLookPath = originalLookPath }()

	cliToolLookPath = func(command string) (string, error) {
		switch command {
		case "ai-memory", "codegraph":
			return "", fmt.Errorf("not found")
		default:
			return "/usr/bin/" + command, nil
		}
	}

	consents := missingInstallConsents([]string{"filesystem", "ai-memory", "codegraph", "unknown"})
	want := []string{"ai-memory", "codegraph"}
	if len(consents) != len(want) {
		t.Fatalf("missingInstallConsents() = %d items, want %d", len(consents), len(want))
	}
	if got := []string{consents[0].Server, consents[1].Server}; !reflect.DeepEqual(got, want) {
		t.Fatalf("missingInstallConsents().servers = %#v, want %#v", got, want)
	}
	if consents[0].Hint == "" || consents[1].Hint == "" {
		t.Fatalf("expected non-empty hints, got %#v", consents)
	}
}

func TestMissingInstallConsentsSkipsInstalledTools(t *testing.T) {
	originalLookPath := cliToolLookPath
	defer func() { cliToolLookPath = originalLookPath }()

	cliToolLookPath = func(command string) (string, error) {
		return "/usr/bin/" + command, nil
	}

	consents := missingInstallConsents([]string{"ai-memory", "codegraph"})
	if len(consents) != 0 {
		t.Fatalf("missingInstallConsents() = %#v, want empty", consents)
	}
}

func TestFormatInstallConsentsDeterministic(t *testing.T) {
	got := formatInstallConsents([]installConsentHint{{
		Server: "ai-memory",
		Hint:   "curl install",
	}, {
		Server: "codegraph",
		Hint:   "bun install",
	}})
	want := []string{"ai-memory: curl install", "codegraph: bun install"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("formatInstallConsents() = %#v, want %#v", got, want)
	}
}

func TestRunPhase4NonInteractiveAllowsDefaults(t *testing.T) {
	plan := &InstallPlan{FilesToInstall: []PlannedFile{}}
	result, action, err := RunPhase4(plan, true, []string{"ai-memory: curl install"})
	if err != nil {
		t.Fatalf("RunPhase4() returned error: %v", err)
	}
	if action != PhaseContinue {
		t.Fatalf("action = %v, want %v", action, PhaseContinue)
	}
	if !result.Confirmed {
		t.Fatalf("result.Confirmed = false, want true")
	}
	if result.InstallConsentsAccepted {
		t.Fatalf("result.InstallConsentsAccepted = true, want false")
	}
}
