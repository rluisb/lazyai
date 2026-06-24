package cmd

import (
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestRequestedReleaseTag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantTag string
	}{
		{name: "plain semver", input: "1.2.3", wantTag: "v1.2.3"},
		{name: "leading v", input: "v1.2.3", wantTag: "v1.2.3"},
		{name: "slash tag", input: "release/train/v1.0.0", wantTag: "release/train/v1.0.0"},
		{name: "rollback tag", input: "pre-refactor-025-phase-2", wantTag: "pre-refactor-025-phase-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestedReleaseTag(tt.input); got != tt.wantTag {
				t.Fatalf("requestedReleaseTag(%q) = %q, want %q", tt.input, got, tt.wantTag)
			}
		})
	}
}

func TestReleaseByTagURLEscapesSlashContainingTags(t *testing.T) {
	got := releaseByTagURL("release/train/v1.0.0")
	want := "https://api.github.com/repos/rluisb/lazyai/releases/tags/release%2Ftrain%2Fv1.0.0"
	if got != want {
		t.Fatalf("releaseByTagURL() = %q, want %q", got, want)
	}
}

func TestFetchReleaseByTagUsesTaggedEndpoint(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", req.Method)
		}
		if req.URL.String() != "https://api.github.com/repos/rluisb/lazyai/releases/tags/release%2Ftrain%2Fv1.0.0" {
			t.Fatalf("url = %s", req.URL.String())
		}
		if got := req.Header.Get("Authorization"); got != "Bearer secret-token" {
			t.Fatalf("authorization = %q, want bearer token", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`{
				"tag_name": "release/train/v1.0.0",
				"assets": []
			}`)),
		}, nil
	})}

	release, err := fetchReleaseByTag(client, "secret-token", "release/train/v1.0.0")
	if err != nil {
		t.Fatalf("fetchReleaseByTag: %v", err)
	}
	if release.TagName != "release/train/v1.0.0" {
		t.Fatalf("TagName = %q", release.TagName)
	}
}

func TestRunUpdateSelfDryRunSpecificTagPreservesExactTag(t *testing.T) {
	origVersion := Version
	origNewAPIClient := newAPIClient
	t.Cleanup(func() {
		Version = origVersion
		newAPIClient = origNewAPIClient
	})

	Version = "v0.9.0"
	mockClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://api.github.com/repos/rluisb/lazyai/releases/tags/release%2Ftrain%2Fv1.0.0" {
			t.Fatalf("url = %s", req.URL.String())
		}
		body := `{
			"tag_name": "release/train/v1.0.0",
			"assets": [
				{"name": "` + binaryAssetName(runtime.GOOS, runtime.GOARCH) + `", "browser_download_url": "https://example.com/binary"},
				{"name": "checksums.txt", "browser_download_url": "https://example.com/checksums"}
			]
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
	newAPIClient = func() *http.Client { return mockClient }

	cmd := newUpdateSelfTestCommand()
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("set dry-run: %v", err)
	}
	if err := cmd.Flags().Set("version", "release/train/v1.0.0"); err != nil {
		t.Fatalf("set version: %v", err)
	}

	stdout, _ := captureOutput(t, func() {
		if err := runUpdateSelf(cmd, nil); err != nil {
			t.Fatalf("runUpdateSelf: %v", err)
		}
	})

	if strings.Contains(stdout, "vrelease/train/v1.0.0") {
		t.Fatalf("stdout corrupts requested tag: %s", stdout)
	}
	if !strings.Contains(stdout, "Would download lazyai-cli release/train/v1.0.0") {
		t.Fatalf("stdout = %q, want exact tag", stdout)
	}
}

func newUpdateSelfTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("check", false, "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("version", "", "")
	return cmd
}
