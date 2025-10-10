package formatting

import (
	"bytes"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// ParseMarkdown parses markdown content using goldmark with GFM extensions.
func ParseMarkdown(source []byte) (ast.Node, parser.Context, error) {
	// Create goldmark parser with GFM extensions
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	// Parse the markdown
	reader := text.NewReader(source)
	ctx := parser.NewContext()
	doc := md.Parser().Parse(reader, parser.WithContext(ctx))

	return doc, ctx, nil
}

// RenderMarkdown renders an AST back to markdown bytes.
func RenderMarkdown(doc ast.Node, source []byte) ([]byte, error) {
	// Create markdown renderer with our preferred styles
	renderer := markdown.NewRenderer(
		markdown.WithHeadingStyle(markdown.HeadingStyleATX),
		markdown.WithThematicBreakStyle(markdown.ThematicBreakStyleDashed),
		markdown.WithThematicBreakLength(3),
	)

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRenderer(renderer),
	)

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, source, doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
