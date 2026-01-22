# Agent Context: Buyruk-CLI

## 🎯 Project Vision
Buyruk is a high-performance, local-first orchestration tool. It treats the filesystem as a database.
- **Core Value**: Near-zero latency, 100% privacy, human-readable JSON storage.
- **Persona**: You are a Senior Go Engineer specializing in systems programming and CLI UX.

## 🛠 Technical Stack
- **Language**: Go 1.2x (Targeting Windows, macOS, Linux)
- **CLI**: `spf13/cobra`
- **TUI/UI**: `charmbracelet/lipgloss`, `olekukonko/tablewriter`, `charmbracelet/glamour`
- **Data**: Local JSON with metadata indexing.

## 💾 Storage & Concurrency (CRITICAL)
**ALL file operations MUST be atomic and race-condition-free.** This is non-negotiable.

### Atomic Write Operations
Every file write MUST follow the atomic safety protocol to prevent corruption:
1. **Check Lock**: Check for `.buyruk.lock`. If present, retry for 5s then fail.
2. **Transaction Log**: Record intent in `.buyruk_pending` before any modification.
3. **Write-Then-Rename**: 
   - Write new data to `filename.json.tmp`.
   - Use `os.Rename` to overwrite `filename.json` (Atomic on Unix/Windows).
4. **Pathing**: NEVER use `/`. Always use `filepath.Join()` to maintain Windows compatibility.
5. **Config Dir**: Use `os.UserConfigDir()` to locate the base `buyruk/` folder.

### Available Atomic Functions (USE THESE)
The storage package provides atomic functions that handle locking and transactions automatically:

- **`WriteJSONAtomic(path, value)`**: Atomic write (overwrites if exists)
  - Use for: Simple writes where overwriting is acceptable
  - Handles: Lock acquisition, transaction, write, commit

- **`WriteJSONAtomicCreate(path, value)`**: Atomic create (fails if exists)
  - Use for: Creating new files (projects, issues, epics)
  - Handles: Lock acquisition, existence check (inside lock), write, commit
  - Returns error if file already exists

- **`UpdateJSONAtomic(path, value, updateFunc)`**: Atomic read-modify-write
  - Use for: Updating existing files (project index, issue updates)
  - Handles: Lock acquisition, read, modify, write, commit
  - Prevents race conditions in read-modify-write operations

### Race Condition Prevention Rules
1. **NEVER check file existence outside a lock** - Use `WriteJSONAtomicCreate` instead
2. **NEVER do read-modify-write without locking** - Use `UpdateJSONAtomic` instead
3. **NEVER use `os.Stat()` + `WriteJSONAtomic()` pattern** - This creates race conditions
4. **Read-only operations** (like `list`) don't need locking - multiple readers are safe

### Examples

**❌ WRONG - Race Condition:**
```go
// Check exists (no lock)
if _, err := os.Stat(path); err == nil {
    return fmt.Errorf("already exists")
}
// Write (with lock, but check was outside lock!)
storage.WriteJSONAtomic(path, data)
```

**✅ CORRECT - Atomic Create:**
```go
// Atomic check + create (all inside lock)
if err := storage.WriteJSONAtomicCreate(path, data); err != nil {
    if strings.Contains(err.Error(), "already exists") {
        return fmt.Errorf("already exists")
    }
    return err
}
```

**❌ WRONG - Race Condition:**
```go
// Read (no lock)
var index models.ProjectIndex
storage.ReadJSON(path, &index)
// Modify
index.AddIssue(issue)
// Write (with lock, but read was outside lock!)
storage.WriteJSONAtomic(path, &index)
```

**✅ CORRECT - Atomic Update:**
```go
// Atomic read-modify-write (all inside lock)
var index models.ProjectIndex
storage.UpdateJSONAtomic(path, &index, func(v interface{}) error {
    idx := v.(*models.ProjectIndex)
    idx.AddIssue(issue)
    return nil
})
```

## 🤖 AI-Native Logic (L-SON)
L-SON is our token-optimized format for LLM context.
- When the `--format lson` flag is used, output key-value pairs prefixed with `@`.
- Example: `@ID: CORE-12`, `@STATUS: TODO`.
- Do not use verbose JSON brackets in L-SON mode.

## 🏗 Directory structure reference
- `config.json`: Global settings.
- `projects/`: Root for all projects.
- `projects/[KEY]/project.json`: The **Index File**. Must be updated whenever an issue is created/deleted.
- `projects/[KEY]/issues/`: Storage for individual task details.

## 🏗 Project Structure (GitHub CLI Pattern)
Following the GitHub CLI (`gh`) architecture pattern:
- **`cmd/buyruk/`**: Entry point (`main.go`) - minimal, just calls `cli.Execute()`
- **`internal/cli/`**: All Cobra commands and command logic
  - `root.go`: Root command with persistent flags
  - `version.go`, `list.go`, etc.: Individual command implementations
- **`internal/build/`**: Build-time metadata (version, build date, etc.)
- **`internal/ui/`**: Rendering logic (tables, colors, markdown)
- **`internal/storage/`**: Filesystem operations and atomic write logic
- **`internal/config/`**: Configuration management

## 🛠 Commands & Workflows
- **Build**: `go build -o buyruk ./cmd/buyruk`
- **Test**: `go test ./...` (Always run tests before suggesting a PR).
- **Repair**: If `project.json` (index) is out of sync with `issues/`, use the repair logic.

## 🧪 Local Development & CI
- **MANDATORY Pre-Commit Checks**: **ALWAYS** run these three commands before ANY commit or push:
  1. **Format**: `gofmt -s -w .` (format all code)
  2. **Vet**: `go vet ./...` (static analysis)
  3. **Test**: `go test ./...` (run all tests)
  - **NEVER skip these checks** - CI will fail if code is not formatted
  - **NEVER commit without running all three** - This is non-negotiable
- **Local First**: All lint and test issues MUST be caught locally before pushing.
  - Run `go test -race ./...` to catch concurrency issues (recommended before PR)
  - Run `go mod verify` to verify dependencies
- **CI Purpose**: CI runs the same checks on multiple OSes (Ubuntu, Windows, macOS) for additional validation. CI should be green if local checks pass.
- **Build Artifacts**: 
  - Artifacts are ONLY uploaded when explicitly requested via:
    1. **PR Label**: Add the `build-artifacts` label to a PR (can be added anytime, even after PR is opened - triggers workflow automatically)
    2. **Manual Trigger**: Use GitHub Actions UI → "Run workflow" → check "Upload build artifacts" checkbox
  - Artifacts are never uploaded automatically, even on push to `main`/`master`
  - Artifacts include binaries for all 3 OSes (Ubuntu, Windows, macOS)
  - Only add the label or use manual trigger when explicitly requested in the prompt
- **WSL & Environment Directives**: 
  - **This project runs in WSL (Windows Subsystem for Linux)**
  - **ALWAYS use `required_permissions: ['all']`** to access the WSL environment directly
  - The sandboxed environment runs as root and lacks SSH keys/credentials and permissions, causing failures operations
  - Using `['all']` permissions allows commands to run as the actual user with proper SSH/GitHub authentication and all other permissions
  - Example: `run_terminal_cmd(..., required_permissions=['all'])` for git operations

## 📝 Rules of Engagement
- **Always Do**: 
  - **MANDATORY: Before EVERY commit/push, run:**
    1. `gofmt -s -w .` (format code)
    2. `go vet ./...` (static analysis)
    3. `go test ./...` (run tests)
    - **This is non-negotiable** - CI will fail if skipped
  - Use early returns. Wrap errors with `fmt.Errorf("context: %w", err)`.
  - Use `cmd.OutOrStdout()` and `cmd.ErrOrStderr()` in Cobra commands (not `os.Stdout`/`os.Stderr`) for testability.
  - Write tests for all new commands and packages.
  - **Create tests for each scenario**: Every function, edge case, error path, and code branch should have a corresponding test. Empty test functions that don't actually test anything should be removed or properly implemented.
  - **Run local checks before suggesting PRs**: `go vet ./...`, `gofmt -s -l .`, `go test -race ./...`, `go mod verify`
  - Ensure CI will be green by catching issues locally first.
- **Ask First**: Before adding new external dependencies to `go.mod`.
- **Never Do**: 
  - Do not use global variables for state; pass a `Context` or `App` struct.
  - Do not use package-level variables for flag storage (causes test pollution and race conditions).
  - Do not use `os.Stdout`/`os.Stderr` directly in command handlers; use Cobra's output writers.
  - Do not modify global state in tests; use `cmd.SetArgs()` to set flags instead.
  - Do not push code that fails local lint/test checks.
  - Do not add `build-artifacts` label to PRs unless explicitly requested.
- **UX**: Errors must go to `os.Stderr` (via `cmd.ErrOrStderr()`). Success messages should use `lipgloss` styles.
- **Command Structure**: 
  - Each command should be in its own file (`internal/cli/<command>.go`)
  - Use `New<Command>Cmd()` factory pattern returning `*cobra.Command`
  - Register commands in `root.go` via `rootCmd.AddCommand()`

## References
Rules in .cursor/rules directory and README.md for project details.