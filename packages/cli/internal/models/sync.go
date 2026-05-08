// Sync logic: pull the live models.dev/api.json catalog, filter to the
// providers our targets care about, and regenerate catalog_gen.go. The
// curated tier lists in catalog.go reference IDs from catalog_gen.go; sync
// also verifies every curated reference still exists upstream so a deleted
// model fails the sync rather than shipping stale curation.

package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"
)

// ModelsDevURL is the canonical catalog endpoint. Swap in tests via the
// fetcher abstraction below — this constant is for production callers.
const ModelsDevURL = "https://models.dev/api.json"

// SyncTargets enumerates the providers that compile-time catalogs reference.
// New providers added to catalog.go should also be added here so sync pulls
// their model list. Order is informational; the generator emits
// alphabetically per-provider for stable output.
var SyncTargets = []string{
	"anthropic",
	"openai",
	"github-copilot",
	"google",
	"ollama-cloud",
	"opencode",
	"opencode-go",
}

// ModelDiff describes one change between the locally-vendored catalog and
// the upstream snapshot. SyncReport.Render prints these grouped by kind.
type ModelDiff struct {
	Kind     string // "added" | "removed" | "changed"
	Provider string
	ID       string
	Note     string // for "changed", a human-readable summary of the delta
}

// SyncReport bundles everything the cmd-level wrapper needs to render a
// preview and decide whether to write. MissingCurated lists curated tier
// references in catalog.go that no longer exist upstream — non-empty
// MissingCurated should fail the sync until the curation is repaired.
type SyncReport struct {
	UpstreamCount   int
	FilteredCount   int
	Diffs           []ModelDiff
	MissingCurated  []string
	GeneratedSource []byte
}

// Fetcher fetches the upstream JSON. Tests inject an in-memory fetcher so
// they can run without network.
type Fetcher func(ctx context.Context) ([]byte, error)

// HTTPFetcher is the production fetcher. 30s is well above the 1.8MB
// payload's transfer time on any reasonable connection but short enough to
// fail fast on a stuck connection.
func HTTPFetcher(ctx context.Context) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, ModelsDevURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models.dev returned %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// Sync pulls the upstream catalog, filters to SyncTargets, diffs against
// the in-memory allModels (the current vendored snapshot), and produces the
// regenerated source. Caller decides whether to write.
func Sync(ctx context.Context, fetch Fetcher) (*SyncReport, error) {
	if fetch == nil {
		fetch = HTTPFetcher
	}
	raw, err := fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch models.dev: %w", err)
	}

	upstream, err := parseModelsDev(raw)
	if err != nil {
		return nil, err
	}

	filtered := filterToTargets(upstream, SyncTargets)
	report := &SyncReport{
		UpstreamCount: len(upstream),
		FilteredCount: len(filtered),
		Diffs:         diffEntries(allModels, filtered),
		MissingCurated: missingCuratedReferences(filtered, []Catalog{
			ClaudeCodeCatalog, OpenCodeCatalog, CopilotCatalog,
		}),
	}
	report.GeneratedSource = generateCatalogGoSource(filtered)
	return report, nil
}

// parseModelsDev decodes the api.json shape into a flat []modelEntry. The
// upstream schema nests models under `<provider>.models.<id>`; we unwrap
// into the same struct catalog_gen.go uses so the diff is straightforward.
func parseModelsDev(raw []byte) ([]modelEntry, error) {
	type modelDoc struct {
		Limit struct {
			Context int `json:"context"`
		} `json:"limit"`
		ToolCall   *bool `json:"tool_call,omitempty"`
		Reasoning  *bool `json:"reasoning,omitempty"`
		Modalities struct {
			Input []string `json:"input,omitempty"`
		} `json:"modalities"`
	}
	type providerDoc struct {
		Models map[string]modelDoc `json:"models"`
	}
	var doc map[string]providerDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse models.dev: %w", err)
	}

	var out []modelEntry
	for prov, p := range doc {
		for id, m := range p.Models {
			reasoning := m.Reasoning != nil && *m.Reasoning
			multimodal := slices.Contains(m.Modalities.Input, "image") ||
				slices.Contains(m.Modalities.Input, "video")
			out = append(out, modelEntry{
				Provider:   prov,
				ID:         id,
				Reasoning:  reasoning,
				Multimodal: multimodal,
				CtxWindow:  m.Limit.Context,
			})
		}
	}
	sortEntries(out)
	return out, nil
}

func filterToTargets(in []modelEntry, targets []string) []modelEntry {
	out := make([]modelEntry, 0, len(in))
	for _, e := range in {
		if slices.Contains(targets, e.Provider) {
			out = append(out, e)
		}
	}
	return out
}

// diffEntries produces a stable ordered list of changes. "changed" fires
// when reasoning, multimodal, or context-window differs by ≥5%. Smaller
// context-window jitter on models.dev's side is ignored to keep churn down.
func diffEntries(local, upstream []modelEntry) []ModelDiff {
	type key struct{ p, id string }
	mk := func(e modelEntry) key { return key{e.Provider, e.ID} }
	localByKey := map[key]modelEntry{}
	for _, e := range local {
		localByKey[mk(e)] = e
	}
	upstreamByKey := map[key]modelEntry{}
	for _, e := range upstream {
		upstreamByKey[mk(e)] = e
	}

	var diffs []ModelDiff
	for k, u := range upstreamByKey {
		l, ok := localByKey[k]
		if !ok {
			diffs = append(diffs, ModelDiff{Kind: "added", Provider: k.p, ID: k.id})
			continue
		}
		var notes []string
		if l.Reasoning != u.Reasoning {
			notes = append(notes, fmt.Sprintf("reasoning %t->%t", l.Reasoning, u.Reasoning))
		}
		if l.Multimodal != u.Multimodal {
			notes = append(notes, fmt.Sprintf("multimodal %t->%t", l.Multimodal, u.Multimodal))
		}
		if ctxJumpExceeds(l.CtxWindow, u.CtxWindow, 0.05) {
			notes = append(notes, fmt.Sprintf("ctx %d->%d", l.CtxWindow, u.CtxWindow))
		}
		if len(notes) > 0 {
			diffs = append(diffs, ModelDiff{Kind: "changed", Provider: k.p, ID: k.id, Note: strings.Join(notes, ", ")})
		}
	}
	for k := range localByKey {
		if _, ok := upstreamByKey[k]; !ok {
			diffs = append(diffs, ModelDiff{Kind: "removed", Provider: k.p, ID: k.id})
		}
	}
	sort.Slice(diffs, func(i, j int) bool {
		if diffs[i].Kind != diffs[j].Kind {
			return diffs[i].Kind < diffs[j].Kind
		}
		if diffs[i].Provider != diffs[j].Provider {
			return diffs[i].Provider < diffs[j].Provider
		}
		return diffs[i].ID < diffs[j].ID
	})
	return diffs
}

func ctxJumpExceeds(a, b int, frac float64) bool {
	if a == 0 || b == 0 {
		return a != b
	}
	delta := float64(a - b)
	if delta < 0 {
		delta = -delta
	}
	return delta/float64(a) > frac
}

// missingCuratedReferences returns curated tier IDs (from catalog.go's
// Frontier/Balanced/Speed lists) that have no matching entry in the
// upstream snapshot. A non-empty result is a hard failure: the sync
// command must not write a regenerated file that strands curated picks.
func missingCuratedReferences(upstream []modelEntry, catalogs []Catalog) []string {
	exists := func(modelField string) bool {
		prov, id := splitProviderModel(modelField)
		for _, e := range upstream {
			if (prov == "" || e.Provider == prov) && e.ID == id {
				return true
			}
		}
		return false
	}
	var missing []string
	seen := map[string]bool{}
	for _, c := range catalogs {
		for _, list := range [][]string{c.Frontier, c.Balanced, c.Speed} {
			for _, m := range list {
				if seen[m] {
					continue
				}
				seen[m] = true
				if !exists(m) {
					missing = append(missing, m)
				}
			}
		}
	}
	sort.Strings(missing)
	return missing
}

// generateCatalogGoSource emits the Go source for catalog_gen.go. The
// header is regenerated each time so the snapshot date stays current; the
// body is sorted (provider, id) for stable output.
func generateCatalogGoSource(entries []modelEntry) []byte {
	sortEntries(entries)
	var b strings.Builder
	b.WriteString("// Code generated by `lazyai models sync` from " + ModelsDevURL + ".\n")
	b.WriteString("// DO NOT EDIT MANUALLY.\n")
	b.WriteString("//\n")
	b.WriteString("// Snapshot: " + time.Now().UTC().Format("2006-01-02") + "\n\n")
	b.WriteString("package models\n\n")
	b.WriteString("type modelEntry struct {\n")
	b.WriteString("\tProvider   string\n")
	b.WriteString("\tID         string\n")
	b.WriteString("\tReasoning  bool\n")
	b.WriteString("\tMultimodal bool\n")
	b.WriteString("\tCtxWindow  int\n")
	b.WriteString("}\n\n")
	b.WriteString("var allModels = []modelEntry{\n")
	currentProvider := ""
	for _, e := range entries {
		if e.Provider != currentProvider {
			if currentProvider != "" {
				b.WriteString("\n")
			}
			b.WriteString("\t// " + e.Provider + "\n")
			currentProvider = e.Provider
		}
		fmt.Fprintf(&b, "\t{%q, %q, %t, %t, %d},\n",
			e.Provider, e.ID, e.Reasoning, e.Multimodal, e.CtxWindow)
	}
	b.WriteString("}\n")
	return []byte(b.String())
}

func sortEntries(es []modelEntry) {
	sort.Slice(es, func(i, j int) bool {
		if es[i].Provider != es[j].Provider {
			return es[i].Provider < es[j].Provider
		}
		return es[i].ID < es[j].ID
	})
}

// Render writes a human-readable summary of the report to w. Used by the
// cmd-level wrapper for both interactive preview and CI logs.
func (r *SyncReport) Render(w io.Writer) {
	fmt.Fprintf(w, "upstream models: %d total, %d after target filter\n",
		r.UpstreamCount, r.FilteredCount)
	if len(r.Diffs) == 0 {
		fmt.Fprintln(w, "no changes since last sync")
	} else {
		fmt.Fprintf(w, "diff: %d entries\n", len(r.Diffs))
		for _, d := range r.Diffs {
			switch d.Kind {
			case "added":
				fmt.Fprintf(w, "  + %s/%s\n", d.Provider, d.ID)
			case "removed":
				fmt.Fprintf(w, "  - %s/%s\n", d.Provider, d.ID)
			case "changed":
				fmt.Fprintf(w, "  ~ %s/%s  (%s)\n", d.Provider, d.ID, d.Note)
			}
		}
	}
	if len(r.MissingCurated) > 0 {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "MISSING CURATED REFERENCES (catalog.go references upstream-removed IDs):")
		for _, m := range r.MissingCurated {
			fmt.Fprintf(w, "  ! %s\n", m)
		}
	}
}
