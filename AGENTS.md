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
Every file write MUST follow the atomic safety protocol to prevent corruption:
1. **Check Lock**: Check for `.buyruk.lock`. If present, retry for 5s then fail.
2. **Transaction Log**: Record intent in `.buyruk_pending` before any modification.
3. **Write-Then-Rename**: 
   - Write new data to `filename.json.tmp`.
   - Use `os.Rename` to overwrite `filename.json` (Atomic on Unix/Windows).
4. **Pathing**: NEVER use `/`. Always use `filepath.Join()` to maintain Windows compatibility.
5. **Config Dir**: Use `os.UserConfigDir()` to locate the base `buyruk/` folder.

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

## 🛠 Commands & Workflows
- **Build**: `go build -o buyruk ./main.go`
- **Test**: `go test ./...` (Always run tests before suggesting a PR).
- **Repair**: If `project.json` (index) is out of sync with `issues/`, use the repair logic.

## 📝 Rules of Engagement
- **Always Do**: Use early returns. Wrap errors with `fmt.Errorf("context: %w", err)`.
- **Ask First**: Before adding new external dependencies to `go.mod`.
- **Never Do**: Do not use global variables for state; pass a `Context` or `App` struct.
- **UX**: Errors must go to `os.Stderr`. Success messages should use `lipgloss` styles.

## References
Rules in .cursor/rules directory and README.md for project details.