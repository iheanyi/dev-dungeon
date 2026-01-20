# Goroutine Race Condition with Struct Fields

## Problem

Tests fail with race detector enabled:

```
WARNING: DATA RACE
Read at 0x00c0003f8c98 by goroutine 140:
  github.com/iheanyi/devdungeon/internal/ui.(*Model).selectMenuItem.func1.1()
      /internal/ui/app.go:867 +0x53

Previous write at 0x00c0003f8c98 by goroutine 139:
  github.com/iheanyi/devdungeon/internal/ui.(*Model).selectMenuItem.func1()
      /internal/ui/app.go:872 +0x13b
```

## Root Cause

A goroutine was spawned that captures a struct field, but the main thread modifies that field before the goroutine executes:

```go
// BUGGY CODE
func() {
    if m.submitDailyCallback != nil && m.pendingSave != nil {
        go func() { _ = m.submitDailyCallback(m.pendingSave) }()  // reads m.pendingSave
    }
    m.pendingSave = nil  // writes m.pendingSave - RACE!
}
```

The goroutine captures `m.pendingSave` by reference. By the time it executes, `m.pendingSave` may have been set to `nil`.

## Solution

Capture the value in a local variable **before** spawning the goroutine:

```go
// FIXED CODE
func() {
    // Capture pendingSave before clearing to avoid race condition
    saveData := m.pendingSave
    if m.submitDailyCallback != nil && saveData != nil {
        go func() { _ = m.submitDailyCallback(saveData) }()  // uses local copy
    }
    m.pendingSave = nil  // safe - goroutine has its own copy
}
```

## Prevention

When spawning goroutines that access struct fields:

1. **Always capture values** in local variables before `go func()`
2. **Run tests with `-race` flag** to catch data races: `go test -race ./...`
3. **Be suspicious of patterns** like:
   ```go
   go func() { ... m.field ... }()
   m.field = ...  // This is a race!
   ```

## Detection

Run tests with race detector:
```bash
go test -race ./...
```

The race detector will report:
- Which goroutines are racing
- Exact file and line numbers
- The read and write operations causing the race

## Related Files

- `internal/ui/app.go:867-872` - Fixed race condition in daily run submission

## Commits

- `56ff0ad` - Fix race condition in daily run submission callback
