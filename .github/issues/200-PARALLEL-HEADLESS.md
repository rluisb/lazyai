# Headless Init Blocks UI and Runs Sequentially

## Problem

During `lazyai-cli init`, the headless populate/init phase runs **sequentially** and **synchronously**, blocking the UI from returning to the user even after "Setup complete!" is shown.

In interactive mode, the user sees the completion banner but the process continues blocking on `opencode run` or `copilot -p` which can hang indefinitely (no real installation, no auth).

## Root Cause

In `cmd/init.go` lines 249-266, the headless init loop is purely sequential:

```go
for _, tool := range ctx.Tools {
    // ... blocking call
    adapt.RunHeadlessInit(adapterCtx, prompt)
}
```

No goroutines, no parallelism.

## Impact

1. **Interactive mode**: User sees "Setup complete!" but CLI is still blocking
2. **Multiple tools**: When using `--tools opencode,claude-code,copilot`, each tool's headless init adds up linearly
3. **No progress feedback**: User doesn't know what's happening during the blocking phase

## Expected Behavior

1. Headless init should run **in parallel** for each tool (using goroutines)
2. UI should return promptly after scaffold completes
3. Long-running headless init should not block the main flow
4. Errors should be aggregated and reported at the end

## Suggested Fix Pattern

```go
var wg sync.WaitGroup
errCh := make(chan error, len(ctx.Tools))

for _, tool := range ctx.Tools {
    wg.Add(1)
    go func(tool string) {
        defer wg.Done()
        // ... tool-specific headless init
    }(tool)
}

wg.Wait()
```

## References

- Related to: #199 verification
- Found during: Interactive wizard testing (Scenario 7)