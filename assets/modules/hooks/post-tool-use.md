---
asset_paths:
    - hooks/postwrite-lint.sh.tmpl
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/post-tool-use.sh
    hook_type: PostToolUse
    timeout: 120
display_name: ✅ post-tool-use
enabled: true
name: post-tool-use
type: hook
---

**Post-write linting and testing hook.** Runs automatically after Claude writes or edits files to catch issues immediately.

This hook is a Go template that generates language-specific linting based on your project:
- **Go**: Runs `golangci-lint` for code quality and `go test` for validation
- **TypeScript/JavaScript**: Executes `npm run lint`/`eslint` and `npm test`/`vitest`
- **Python**: Runs `ruff` for linting and `pytest` for tests
- **Rust**: Executes `cargo clippy` and `cargo test`
- **C++**: Runs `clang-tidy` and `cppcheck` on recent files
- **Java/Kotlin**: Executes Gradle/Maven build and test tasks
- **PHP**: Validates syntax with `php -l`, runs `phpcs` and `phpunit`
- **Ruby**: Executes `rubocop` and `rspec`
- **Swift**: Runs `swift test` and `swiftlint`
- **C#**: Executes `dotnet build` and `dotnet test`
- And many more languages...

All commands use `|| true` to never block Claude - they provide feedback without stopping the workflow. This gives you immediate validation feedback while keeping Claude's responses flowing.