package render

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Renderer converts markdown to HTML fragments.
type Renderer struct {
	md goldmark.Markdown
}

// New creates a Renderer with GFM extensions and link rewriting enabled.
// linkBase is the URL prefix for wiki links, e.g. "/wiki/".
func New(linkBase string) *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(html.WithUnsafe()),
		goldmark.WithParserOptions(
			parser.WithASTTransformers(
				util.Prioritized(&linkRewriter{linkBase: linkBase}, 100),
			),
		),
	)
	return &Renderer{md: md}
}

// RenderMarkdown converts markdown bytes to a safe HTML fragment.
func (r *Renderer) RenderMarkdown(src []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := r.md.Convert(src, &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// linkRewriter is a goldmark AST transformer that rewrites .md links to HTTP paths.
type linkRewriter struct {
	linkBase string
}

func (l *linkRewriter) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Link:
			dest := string(v.Destination)
			if rewritten, ok := rewriteLink(dest, l.linkBase); ok {
				v.Destination = []byte(rewritten)
			}
		case *ast.Image:
			dest := string(v.Destination)
			// Images in attachments are served from /attachments/.
			if strings.HasPrefix(dest, "../attachments/") || strings.HasPrefix(dest, "attachments/") {
				name := dest[strings.LastIndex(dest, "/")+1:]
				v.Destination = []byte("/attachments/" + name)
			}
		}
		return ast.WalkContinue, nil
	})
}

// rewriteLink converts a relative .md link to an HTTP path.
// Returns the rewritten path and true if a rewrite was performed.
func rewriteLink(dest, linkBase string) (string, bool) {
	// Skip absolute URLs and anchor-only links.
	if strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://") ||
		strings.HasPrefix(dest, "#") || strings.HasPrefix(dest, "/") {
		return dest, false
	}

	// Strip any leading path components (wiki pages are flat).
	name := dest
	if idx := strings.LastIndex(dest, "/"); idx >= 0 {
		name = dest[idx+1:]
	}

	if strings.HasSuffix(name, ".md") {
		page := strings.TrimSuffix(name, ".md")
		return linkBase + page, true
	}

	return dest, false
}
