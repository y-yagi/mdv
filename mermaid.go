package main

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type mermaidExtension struct{}

func (e *mermaidExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newMermaidRenderer(), 100),
	))
}

type mermaidRenderer struct {
	html.Config
}

func newMermaidRenderer() renderer.NodeRenderer {
	return &mermaidRenderer{Config: html.NewConfig()}
}

func (r *mermaidRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *mermaidRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	language := n.Language(source)

	if string(language) != "mermaid" {
		return r.renderDefault(w, source, n, entering, language)
	}

	if entering {
		_, _ = w.WriteString(`<pre class="mermaid">`)
		writeLines(r.Writer, w, source, n)
	} else {
		_, _ = w.WriteString("</pre>\n")
	}
	return ast.WalkContinue, nil
}

func (r *mermaidRenderer) renderDefault(w util.BufWriter, source []byte, n *ast.FencedCodeBlock, entering bool, language []byte) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<pre><code")
		if language != nil {
			_, _ = w.WriteString(" class=\"language-")
			r.Writer.Write(w, language)
			_, _ = w.WriteString("\"")
		}
		_ = w.WriteByte('>')
		writeLines(r.Writer, w, source, n)
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func writeLines(writer html.Writer, w util.BufWriter, source []byte, n *ast.FencedCodeBlock) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		writer.RawWrite(w, line.Value(source))
	}
}
