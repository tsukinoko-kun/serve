package md

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var (
	parserExtensions = parser.Tables |
		parser.FencedCode |
		parser.Autolink |
		parser.Strikethrough |
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

	rendererOptions = html.RendererOptions{Flags: html.CommonFlags | html.NoreferrerLinks | html.NoopenerLinks | html.LazyLoadImages}
)

func WriteMarkdown(w http.ResponseWriter, r *http.Request, title string, md []byte, dirPath string) error {
	p := parser.NewWithExtensions(parserExtensions)
	renderer := html.NewRenderer(rendererOptions)

	doc := p.Parse(md)

	htmlBody := markdown.Render(doc, renderer)

	return WriteDoc(w, r, title, dirPath, htmlBody,
		`<script id="MathJax-script" async defer src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js"></script>`,
		`<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/styles/github-dark.min.css">`,
		`<script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/highlight.min.js"></script>`,
		`<script defer>hljs.highlightAll();</script>`)
}

const (
	htmlDocHeadStr = `<!DOCTYPE html>` +
		`<html>` +
		`<head>` +
		`<meta charset="utf-8">` +
		`<meta name="viewport" content="width=device-width, initial-scale=1">` +
		`<style>` +
		`:root {` +
		`color-scheme: light dark;` +
		`}` +
		`* {` +
		`min-width: 0;` +
		`min-height: 0;` +
		`box-sizing: border-box;` +
		`}` +
		`body {` +
		`font-family: sans-serif;` +
		`}` +
		`h1, h2, h3, h4, h5, h6 {` +
		`font-family: serif;` +
		`}` +
		`code, pre {` +
		`font-family: monospace;` +
		`}` +
		`main {` +
		`max-width: 1000px;` +
		`margin: 0 auto;` +
		`padding: 4rem 1rem;` +
		`}` +
		`a {` +
		`color: #007bff;` +
		`text-decoration: underline;` +
		`}` +
		`a:hover {` +
		`color: #0056b3;` +
		`}` +
		`table {` +
		`border-collapse: collapse;` +
		`width: 100%;` +
		`}` +
		`table, td, th {` +
		`border: 1px solid;` +
		`}` +
		`td, th {` +
		`padding: 0.5rem;` +
		`}` +
		`blockquote {` +
		`margin: 0;` +
		`padding: 0.5rem;` +
		`background-color: #007bff40;` +
		`border-left: 0.25rem solid #007bff;` +
		`}` +
		`code, blockquote {` +
		`border-radius: 0.25rem;` +
		`}` +
		`pre > code:not(.hljs) {` +
		`color: rgb(152, 168, 222);` +
		`background-color: rgb(22, 25, 29);` +
		`padding: 1em;` +
		`display: block;` +
		`}` +
		`</style>`
	htmlDocUpdateScriptStr = `<script defer>` +
		`const hashEl = document.querySelector('meta[name="serve-hash"]');` +
		`if (!hashEl) {` +
		`throw new Error("serve-hash meta tag not found");` +
		`}` +
		`const hash = hashEl.content;` +
		`async function isContentUpToDate() {` +
		`const response = await fetch(window.location.href, {cache: "no-store", method: "HEAD"});` +
		`const newHash = response.headers.get("Serve-Hash");` +
		`if (newHash !== hash) {` +
		`window.location.reload();` +
		`}` +
		`}` +
		`window.setInterval(isContentUpToDate, 2000);` +
		`</script>`
	anchorsScript = `<script defer>` +
		`console.log(dirPath);` +
		`const anchors = document.querySelectorAll("a[href]");` +
		`for(const a of anchors) {` +
		`const href = a.getAttribute("href");` +
		`if (href.startsWith(".")) {` +
		`const url = new URL(window.location);` +
		`url.pathname = ((dirPath || url.pathname) + href.substring(1)).replace(/\/+/g,"/");` +
		`a.setAttribute("href", url.href);` +
		`}else{` +
		`a.setAttribute("target", "_blank");` +
		`}` +
		`}` +
		`</script>`
)

func doc(title string, body []byte, dirPath string, libs ...string) []byte {
	htmlDoc := strings.Builder{}
	htmlDoc.WriteString(htmlDocHeadStr)
	htmlDoc.WriteString(`<meta name="serve-hash" content="`)
	htmlDoc.WriteString(contentHash(body))
	htmlDoc.WriteString(`">`)
	htmlDoc.WriteString(fmt.Sprintf(`<script>const dirPath = %q;</script>`, dirPath))
	htmlDoc.WriteString("<title>")
	htmlDoc.WriteString(title)
	htmlDoc.WriteString("</title></head><body><main>")
	htmlDoc.Write(body)
	htmlDoc.WriteString("</main>")

	for _, lib := range libs {
		htmlDoc.WriteString(lib)
	}

	htmlDoc.WriteString(htmlDocUpdateScriptStr)
	htmlDoc.WriteString(anchorsScript)
	htmlDoc.WriteString("</body></html>")

	return []byte(htmlDoc.String())
}

var h = sha1.New()

// returns the SHA-1 hash of the content as base64 encoded string
func contentHash(content []byte) string {
	return base64.StdEncoding.EncodeToString(h.Sum(content))
}

func WriteDoc(w http.ResponseWriter, r *http.Request, title string, dirPath string, body []byte, libs ...string) error {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Serve-Hash", contentHash(body))

	if r.Method == http.MethodHead {
		return nil
	}

	_, err := w.Write(doc(title, body, dirPath, libs...))
	return err
}
