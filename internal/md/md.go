package md

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func Compile(title string, md []byte) []byte {
	extensions :=
		parser.Tables |
			parser.FencedCode |
			parser.Autolink |
			parser.Strikethrough |
			parser.SpaceHeadings |
			parser.Footnotes |
			parser.NoEmptyLineBeforeBlock |
			parser.HeadingIDs |
			parser.Titleblock |
			parser.AutoHeadingIDs |
			parser.DefinitionLists |
			parser.MathJax |
			parser.OrderedListStart |
			parser.Attributes |
			parser.SuperSubscript |
			parser.EmptyLinesBreakList |
			parser.Includes

	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.NoreferrerLinks | html.NoopenerLinks | html.LazyLoadImages
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlBody := markdown.Render(doc, renderer)

	return Doc(title, htmlBody,
		"<script id=\"MathJax-script\" async defer src=\"https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js\"></script>",
		"<link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/styles/github-dark.min.css\">",
		"<script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/highlight.min.js\"></script>",
		"<script defer>hljs.highlightAll();</script>")
}

func Doc(title string, body []byte, libs ...string) []byte {
	htmlDoc := strings.Builder{}
	htmlDoc.WriteString("<!DOCTYPE html>\n")
	htmlDoc.WriteString("<html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">")
	htmlDoc.WriteString("<style>:root { color-scheme: light dark; } body { font-family: sans-serif; } h1, h2, h3, h4, h5, h6 { font-family: sans-serif; }</style>")
	htmlDoc.WriteString("<title>")
	htmlDoc.WriteString(title)
	htmlDoc.WriteString("</title>")
	htmlDoc.WriteString("</head><body>")
	htmlDoc.Write(body)

	for _, lib := range libs {
		htmlDoc.WriteString(lib)
	}

	htmlDoc.WriteString("</body></html>")

	return []byte(htmlDoc.String())
}
