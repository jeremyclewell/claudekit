package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateSubagentAssetFile creates a subagent markdown file.
func GenerateSubagentAssetFile(desc AssetFileDescriptor, outputPath string) GenerationResult {
	// Use module description if available, otherwise generate placeholder
	var shortDesc, fullDesc, instructions string
	var tools []string

	if desc.Module != nil && desc.Module.GetDescription() != "" {
		// Extract short description for frontmatter (first line or sentence)
		fullDesc = desc.Module.GetDescription()
		lines := strings.Split(strings.TrimSpace(fullDesc), "\n")
		if len(lines) > 0 {
			// Take first line, strip markdown headers
			shortDesc = strings.TrimPrefix(strings.TrimPrefix(lines[0], "##"), "#")
			shortDesc = strings.TrimSpace(shortDesc)
		}

		// Extract tools based on category
		tools = GetToolsForAgent(desc.Name, desc.Module.GetCategory())
		instructions = GenerateInstructionsForAgent(desc.Name, desc.Module.GetCategory())
	} else {
		// Placeholder content
		shortDesc = fmt.Sprintf("TODO: Brief description for %s", desc.Name)
		fullDesc = fmt.Sprintf("TODO: Describe %s agent role and capabilities", desc.Name)
		tools = []string{"Read", "Write", "Edit", "Grep", "Bash"}
		instructions = fmt.Sprintf("TODO: Define workflow for %s:\n1. Step 1\n2. Step 2\n3. Step 3", desc.Name)
	}

	// Build YAML frontmatter
	toolsList := strings.Join(tools, ", ")

	// Build markdown content
	content := fmt.Sprintf(`---
name: %s
description: %s
tools: %s
---

%s

## Instructions

%s

## Examples

%s
`, desc.Name, shortDesc, toolsList, fullDesc, instructions, GenerateExamplesMarkdown(desc.Name))

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return GenerationResult{
			FilePath: outputPath,
			Status:   StatusFailed,
			Error:    fmt.Errorf("failed to create directory: %w", err),
		}
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return GenerationResult{
			FilePath: outputPath,
			Status:   StatusFailed,
			Error:    fmt.Errorf("failed to write file: %w", err),
		}
	}

	isPlaceholder := desc.Module == nil || desc.Module.GetDescription() == ""
	status := StatusSuccess
	if isPlaceholder {
		status = StatusPlaceholderGenerated
	}

	return GenerationResult{
		FilePath:      outputPath,
		Status:        status,
		BytesWritten:  len(content),
		IsPlaceholder: isPlaceholder,
	}
}

// GetToolsForAgent returns appropriate tools based on agent type.
func GetToolsForAgent(name, category string) []string {
	baseTools := []string{"Read", "Write", "Edit", "Grep"}

	switch {
	case strings.Contains(name, "test"):
		return append(baseTools, "Bash", "Task")
	case strings.Contains(name, "security"):
		return append(baseTools, "Bash", "WebSearch", "Task")
	case strings.Contains(name, "perf"):
		return append(baseTools, "Bash", "Task")
	case strings.Contains(name, "data"):
		return append(baseTools, "Bash", "Task")
	case strings.Contains(name, "release"):
		return append(baseTools, "Bash", "Task")
	default:
		return append(baseTools, "Bash", "Task")
	}
}

// GenerateInstructionsForAgent creates workflow instructions.
func GenerateInstructionsForAgent(name, category string) string {
	switch {
	case strings.Contains(name, "code-review"):
		return `### Review Process

1. **Context Gathering**
   - Run git diff to identify changed files and scope
   - Use git log --oneline -5 to understand recent development context
   - Read related files to understand broader impact
   - Check if changes affect public APIs, data models, or critical paths

2. **Review Categories**

**CRITICAL ISSUES** (Must fix before merge):
   - Security vulnerabilities (injection, XSS, auth bypass)
   - Memory leaks, race conditions, deadlocks
   - Breaking changes to public APIs without versioning
   - Data corruption risks or unsafe operations

**WARNINGS** (Should fix):
   - Performance anti-patterns (N+1 queries, inefficient algorithms)
   - Code smells (large functions, deep nesting, duplicated logic)
   - Missing error handling or inadequate logging
   - Inconsistent patterns or style violations

**SUGGESTIONS** (Nice to have):
   - Refactoring opportunities for better readability
   - More descriptive naming or documentation
   - Alternative approaches or libraries

3. **Provide Actionable Feedback**
   - Reference specific file names and line numbers
   - Explain *why* something is problematic
   - Suggest concrete alternatives with code examples
   - Prioritize feedback by severity`

	case strings.Contains(name, "test"):
		return `### Testing Workflow

1. **Discover Test Framework**
   - Check for pytest, jest, go test, cargo test, etc.
   - Identify test file patterns (*_test.go, *.test.js, etc.)
   - Read test configuration files

2. **Run Tests**
   - Execute full test suite: npm test, go test ./..., pytest
   - Run specific test files if debugging: go test -run TestName
   - Check for test coverage: go test -cover, pytest --cov

3. **Analyze Failures**
   - Read error messages and stack traces carefully
   - Identify patterns in failures (timing, environment, data)
   - Check if failures are flaky or deterministic
   - Use debugging tools if needed (dlv, pdb, Chrome DevTools)

4. **Fix or Improve Tests**
   - Fix broken test assertions
   - Add missing test cases for edge conditions
   - Improve test clarity and maintainability
   - Ensure tests are isolated and don't depend on order`

	case strings.Contains(name, "bug"):
		return `### RIDDE Debugging Methodology

1. **Reproduce** the issue
   - Get exact steps to reproduce
   - Identify minimal reproduction case
   - Note environment details (OS, version, config)
   - Try to reproduce locally

2. **Investigate** symptoms
   - Check logs for errors and warnings
   - Add strategic logging/print statements
   - Use debugger to inspect state (dlv, pdb, Chrome DevTools)
   - Review recent changes with git log and git blame

3. **Deduce** root cause
   - Form hypotheses about the cause
   - Test each hypothesis systematically
   - Trace execution flow through code
   - Check for common patterns: race conditions, null pointers, type mismatches

4. **Document** findings
   - Write clear description of root cause
   - Note why the bug occurred
   - Document steps taken to identify it
   - Include relevant code snippets

5. **Execute** solution
   - Implement minimal fix
   - Add regression test
   - Verify fix resolves original issue
   - Check for similar bugs elsewhere in codebase`

	case strings.Contains(name, "security"):
		return `### Security Audit Process

1. **OWASP Top 10 Scan**
   - **Injection**: Check for SQL, NoSQL, command, LDAP injection
   - **Broken Authentication**: Review session management, password policies
   - **Sensitive Data Exposure**: Check encryption at rest and in transit
   - **XML External Entities**: Review XML parsers for XXE vulnerabilities
   - **Broken Access Control**: Verify authorization checks
   - **Security Misconfiguration**: Review headers, CORS, error messages
   - **XSS**: Check for reflected, stored, and DOM-based XSS
   - **Insecure Deserialization**: Review serialization libraries
   - **Components with Known Vulnerabilities**: Scan dependencies
   - **Insufficient Logging**: Verify security event logging

2. **Dependency Analysis**
   - Run npm audit, go mod verify, or safety check
   - Check for outdated packages with known CVEs
   - Review transitive dependencies
   - Suggest version upgrades or patches

3. **Authentication & Authorization**
   - Review authentication mechanisms
   - Check for proper password hashing (bcrypt, argon2)
   - Verify JWT implementation and secret management
   - Test authorization boundaries (can user A access user B's data?)

4. **Provide Remediation**
   - Reference specific OWASP guidelines
   - Provide code examples for fixes
   - Suggest security libraries and best practices
   - Recommend automated security scanning tools`

	case strings.Contains(name, "perf"):
		return `### Performance Optimization Workflow

1. **Establish Baseline**
   - Run benchmarks to establish current performance
   - Use profiling tools: go test -bench, perf, Chrome DevTools
   - Identify performance goals (latency, throughput, memory)

2. **Profile and Identify Bottlenecks**
   - CPU profiling: Find hot code paths
   - Memory profiling: Identify allocations and leaks
   - I/O profiling: Check disk and network operations
   - Database profiling: Use EXPLAIN for slow queries

3. **Analyze Bottlenecks**
   - **Algorithm complexity**: Is O(n²) algorithm causing slowdown?
   - **Database issues**: N+1 queries, missing indexes, large result sets?
   - **Memory issues**: Unnecessary allocations, large objects in memory?
   - **I/O bottlenecks**: Synchronous operations blocking?

4. **Optimize and Measure**
   - Implement targeted optimizations
   - Re-run benchmarks to measure improvement
   - Ensure optimizations don't harm readability
   - Document performance characteristics

5. **Suggest Architectural Improvements**
   - Caching strategies (Redis, in-memory)
   - Database indexing and query optimization
   - Async/parallel processing opportunities
   - Load balancing and horizontal scaling`

	case strings.Contains(name, "docs"):
		return `### Documentation Workflow

1. **Read and Understand Code**
   - Read the code thoroughly to understand functionality
   - Identify public APIs, interfaces, and entry points
   - Note complex algorithms or non-obvious logic
   - Check existing documentation for gaps

2. **Generate Documentation**
   - **README**: Project overview, installation, quick start
   - **API Docs**: Function signatures, parameters, return values
   - **Guides**: How-to guides for common tasks
   - **Architecture**: System design, component relationships
   - **Examples**: Code samples showing real usage

3. **Follow Best Practices**
   - Use clear, concise language
   - Include code examples that actually work
   - Document edge cases and limitations
   - Add diagrams for complex flows (mermaid, PlantUML)
   - Keep docs in sync with code

4. **Update Existing Documentation**
   - Mark deprecated APIs
   - Update changed behavior
   - Fix broken examples
   - Add migration guides for breaking changes`

	case strings.Contains(name, "release"):
		return `### Release Preparation Workflow

1. **Version Verification**
   - Check version numbers in: package.json, go.mod, Cargo.toml, etc.
   - Ensure semantic versioning (MAJOR.MINOR.PATCH)
   - Update version in all relevant files

2. **Generate Changelog**
   - Run git log to review commits since last release
   - Group changes by type: Features, Fixes, Breaking Changes
   - Use conventional commits format if available
   - Highlight notable changes and migration steps

3. **Pre-Release Checks**
   - Run full test suite: npm test, go test ./..., cargo test
   - Run linters and formatters
   - Build production artifacts
   - Check for uncommitted changes
   - Review security vulnerabilities

4. **Create Release**
   - Tag release: git tag -a v1.2.3 -m "Release v1.2.3"
   - Push tags: git push --tags
   - Create GitHub release with changelog
   - Publish packages: npm publish, cargo publish

5. **Post-Release**
   - Verify package published successfully
   - Update documentation with new version
   - Announce release if applicable
   - Monitor for issues in production`

	case strings.Contains(name, "data"):
		return `### Data Analysis Workflow

1. **Understand Data Schema**
   - Query schema: SHOW TABLES, DESCRIBE table
   - Identify primary/foreign keys and relationships
   - Check data types and constraints
   - Review indexes and performance characteristics

2. **Exploratory Analysis**
   - Get row counts: SELECT COUNT(*) FROM table
   - Check data distribution: MIN, MAX, AVG, percentiles
   - Identify null values and data quality issues
   - Sample data to understand patterns

3. **Generate SQL Queries**
   - Write optimized SELECT queries
   - Use proper JOINs (INNER, LEFT, RIGHT)
   - Add WHERE clauses for filtering
   - Use GROUP BY for aggregations
   - Apply HAVING for filtered aggregations
   - Order and limit results appropriately

4. **Data Transformations**
   - Clean and normalize data
   - Handle missing values
   - Convert data types as needed
   - Create derived columns

5. **Provide Insights**
   - Summarize key findings
   - Identify trends and anomalies
   - Suggest data quality improvements
   - Recommend indexes for common queries`

	default:
		return fmt.Sprintf("TODO: Define workflow for %s:\n1. Step 1\n2. Step 2\n3. Step 3", name)
	}
}

// GenerateExamplesMarkdown creates usage examples in markdown format.
func GenerateExamplesMarkdown(name string) string {
	examples := GenerateExamplesForAgent(name)
	var result strings.Builder
	for _, example := range examples {
		result.WriteString(fmt.Sprintf("- %s\n", example))
	}
	return result.String()
}

// GenerateExamplesForAgent creates usage examples.
func GenerateExamplesForAgent(name string) []string {
	switch {
	case strings.Contains(name, "code-review"):
		return []string{
			"Review the authentication logic in src/auth.go",
			"Check the API endpoints for security best practices",
		}
	case strings.Contains(name, "test"):
		return []string{
			"Run the test suite and fix any failing tests",
			"Analyze test coverage and suggest missing test cases",
		}
	case strings.Contains(name, "security"):
		return []string{
			"Audit the codebase for OWASP Top 10 vulnerabilities",
			"Check dependencies for known security issues",
		}
	case strings.Contains(name, "bug"):
		return []string{
			"Debug the issue reported in GitHub issue #123",
			"Find the root cause of the intermittent database connection error",
		}
	default:
		return []string{fmt.Sprintf("TODO: Add example usage scenario for %s", name)}
	}
}

// GeneratePlaceholderSubagent creates a placeholder subagent JSON file.
func GeneratePlaceholderSubagent(name string, outputPath string) GenerationResult {
	desc := AssetFileDescriptor{Name: name, Type: AssetTypeSubagent}
	return GenerateSubagentAssetFile(desc, outputPath)
}
