# claudekit - Complete Project Demo
# Interactive TUI demonstration showing the full project setup workflow

Output demo.gif

Set Shell bash
Set FontSize 14
Set Width 1600
Set Height 1000
Set Theme "Catppuccin Mocha"
Set Padding 20

# Set truecolor mode for gradient rendering
Env COLORTERM truecolor

# Launch the application
Type "./claudekit"
Enter
Sleep 2s

# ============================================
# Form Interaction - Fill out the configuration
# ============================================

# Move to Description field
Tab
Sleep 300ms
Type "A sample Claude Code project with hooks, agents, and commands"
Sleep 500ms

# Move to Languages field
Tab
Sleep 300ms
Type "Go"
Sleep 300ms
Down
Sleep 300ms
Type " "  # Space to toggle selection
Sleep 500ms

# Add TypeScript
Down
Down
Down
Down
Sleep 300ms
Type " "  # Toggle TypeScript
Sleep 500ms

# Add Python
Down
Sleep 300ms
Type " "  # Toggle Python
Sleep 1s

# Move to Scope selector (Local/Project)
Tab
Sleep 300ms
# Select Project scope
Right
Sleep 500ms

# ============================================
# Subagents Selection
# ============================================
Tab
Sleep 300ms

# Select code-reviewer
Type " "
Sleep 300ms

# Move down and select test-runner
Down
Sleep 300ms
Type " "
Sleep 300ms

# Select bug-sleuth
Down
Sleep 300ms
Type " "
Sleep 300ms

# Select security-auditor
Down
Down
Sleep 300ms
Type " "
Sleep 1s

# ============================================
# Hooks Selection
# ============================================
Tab
Sleep 300ms

# Select session-start
Type " "
Sleep 300ms

# Select pre-tool-use (prewrite guard)
Down
Sleep 300ms
Type " "
Sleep 300ms

# Select post-tool-use (postwrite lint)
Down
Sleep 300ms
Type " "
Sleep 300ms

# Select user-prompt-submit
Down
Sleep 300ms
Type " "
Sleep 1s

# ============================================
# Commands Selection
# ============================================
Tab
Sleep 300ms

# Select add-feature
Type " "
Sleep 300ms

# Select debug-issue
Down
Down
Sleep 300ms
Type " "
Sleep 300ms

# Select fix-github-issue
Down
Sleep 300ms
Type " "
Sleep 1s

# ============================================
# MCP Servers Selection
# ============================================
Tab
Sleep 300ms

# Select filesystem
Type " "
Sleep 300ms

# Select github
Down
Sleep 300ms
Type " "
Sleep 300ms

# Select postgres
Down
Down
Down
Sleep 300ms
Type " "
Sleep 1.5s

# ============================================
# Generate Configuration
# ============================================

# Move to Generate button
Tab
Tab
Sleep 500ms

# Press Generate
Enter
Sleep 2s

# ============================================
# View Configuration Summary
# ============================================

# Scroll through the summary
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 2s

# ============================================
# Exit Application
# ============================================

Enter
Sleep 1s

# Show generated files
Type "ls -la .claude/"
Enter
Sleep 1.5s

Type "ls -la .claude/hooks/"
Enter
Sleep 1.5s

Type "ls -la .claude/agents/"
Enter
Sleep 1.5s

# Show CLAUDE.md was created
Type "head -20 CLAUDE.md"
Enter
Sleep 2s

# Show settings.json structure
Type "cat .claude/settings.json | head -30"
Enter
Sleep 2s
