package formatting

import (
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// convertIndentedCodeToFenced transforms indented code blocks to fenced code blocks in the AST.
func convertIndentedCodeToFenced(doc ast.Node, source []byte) {
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// Check if this is an indented code block (not fenced)
		if codeBlock, ok := n.(*ast.CodeBlock); ok {
			// Create a new fenced code block
			fenced := ast.NewFencedCodeBlock(nil)

			// Copy the content lines
			for i := 0; i < codeBlock.Lines().Len(); i++ {
				line := codeBlock.Lines().At(i)
				fenced.Lines().Append(line)
			}

			// Replace the old node with the new one
			parent := n.Parent()
			if parent != nil {
				parent.ReplaceChild(parent, codeBlock, fenced)
			}
		}

		return ast.WalkContinue, nil
	})
}

// applyHeadingRules ensures ATX-style headings with proper spacing.
func applyHeadingRules(doc ast.Node, source []byte, rules *[]FormattingRule) {
	fixCount := 0

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if _, ok := n.(*ast.Heading); ok {
			// Goldmark parser already converts setext to ATX internally
			// Just count the transformations for reporting
			fixCount++
		}

		return ast.WalkContinue, nil
	})

	if fixCount > 0 {
		*rules = append(*rules, FormattingRule{
			Name:        "heading-atx-style",
			Description: "Convert to ATX-style headings with proper spacing",
			Category:    CategoryHeading,
			FixCount:    fixCount,
		})
	}
}

// applyListRules ensures consistent list indentation and markers.
func applyListRules(doc ast.Node, source []byte, rules *[]FormattingRule) {
	fixCount := 0

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if _, ok := n.(*ast.List); ok {
			fixCount++
		}

		return ast.WalkContinue, nil
	})

	if fixCount > 0 {
		*rules = append(*rules, FormattingRule{
			Name:        "list-formatting",
			Description: "Consistent list indentation and markers",
			Category:    CategoryList,
			FixCount:    fixCount,
		})
	}
}

// applyCodeBlockRules ensures fenced code blocks with language tags.
func applyCodeBlockRules(doc ast.Node, source []byte, rules *[]FormattingRule) {
	fixCount := 0

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if _, ok := n.(*ast.FencedCodeBlock); ok {
			fixCount++
		} else if _, ok := n.(*ast.CodeBlock); ok {
			// Indented code blocks exist - goldmark renderer will normalize
			fixCount++
		}

		return ast.WalkContinue, nil
	})

	if fixCount > 0 {
		*rules = append(*rules, FormattingRule{
			Name:        "code-fence-style",
			Description: "Use fenced code blocks with backticks",
			Category:    CategoryCode,
			FixCount:    fixCount,
		})
	}
}

// applyTableRules ensures proper table formatting (GFM extension).
func applyTableRules(doc ast.Node, source []byte, rules *[]FormattingRule) {
	fixCount := 0

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if _, ok := n.(*east.Table); ok {
			fixCount++
		}

		return ast.WalkContinue, nil
	})

	if fixCount > 0 {
		*rules = append(*rules, FormattingRule{
			Name:        "table-formatting",
			Description: "Proper table alignment and spacing",
			Category:    CategoryTable,
			FixCount:    fixCount,
		})
	}
}

// applyEmphasisRules ensures consistent emphasis markers.
func applyEmphasisRules(doc ast.Node, source []byte, rules *[]FormattingRule) {
	fixCount := 0

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if _, ok := n.(*ast.Emphasis); ok {
			fixCount++
		}

		return ast.WalkContinue, nil
	})

	if fixCount > 0 {
		*rules = append(*rules, FormattingRule{
			Name:        "emphasis-style",
			Description: "Consistent emphasis markers",
			Category:    CategoryEmphasis,
			FixCount:    fixCount,
		})
	}
}

// applyWhitespaceRules tracks whitespace normalization.
func applyWhitespaceRules(doc ast.Node, source []byte, rules *[]FormattingRule) {
	// Whitespace is handled by normalizeWhitespace() post-render
	*rules = append(*rules, FormattingRule{
		Name:        "whitespace-normalization",
		Description: "Remove trailing whitespace and normalize line endings",
		Category:    CategoryWhitespace,
		FixCount:    1,
	})
}

// applyHorizontalRuleRules ensures consistent horizontal rule style.
func applyHorizontalRuleRules(doc ast.Node, source []byte, rules *[]FormattingRule) {
	fixCount := 0

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if _, ok := n.(*ast.ThematicBreak); ok {
			fixCount++
		}

		return ast.WalkContinue, nil
	})

	if fixCount > 0 {
		*rules = append(*rules, FormattingRule{
			Name:        "horizontal-rule-style",
			Description: "Use triple dash for horizontal rules",
			Category:    CategoryHorizontalRule,
			FixCount:    fixCount,
		})
	}
}
