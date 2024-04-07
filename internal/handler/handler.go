package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	posixpath "path"
	"path/filepath"
	"strings"

	"github.com/Frank-Mayer/serve/internal/md"
	"github.com/Frank-Mayer/serve/internal/utils"
	"github.com/charmbracelet/log"
)

var (
	fileServer http.Handler
	absPath    string
	mdCompile  bool
)

func Init(dirPath string, compileMd bool) {
	absPath = dirPath
	mdCompile = compileMd
	fileServer = http.FileServer(http.Dir(absPath))
}

func httpError(w http.ResponseWriter, error string, code int) {
	log.Error("error during request", "code", code, "error", error)
	http.Error(w, error, code)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	reqPath := filepath.Join(absPath, r.URL.Path)

	defer func() {
		if rec := recover(); rec != nil {
			httpError(w, fmt.Sprintf("panic during request to %s: %v", r.URL, rec), http.StatusInternalServerError)
		} else {
			log.Debug("Request", "url", r.URL, "path", reqPath, "method", r.Method, "remote", r.RemoteAddr)
		}
	}()

	if err := handleFileExistence(w, r, reqPath); err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if ok, err := handleDirectory(w, r, reqPath); err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	} else if ok {
		return
	}

	if mdCompile {
		if err := handleMarkdown(w, r, reqPath); err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// If none of the special cases were met, serve the file
	fileServer.ServeHTTP(w, r)
}

func handleFileExistence(_ http.ResponseWriter, r *http.Request, reqPath string) error {
	_, err := os.Stat(reqPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %q does not exist", r.URL.Path)
		}
		return fmt.Errorf("failed to check if file %q exists: %v", r.URL.Path, err)
	}

	if !utils.IsIn(reqPath, absPath) {
		return fmt.Errorf("file %q does not exist", r.URL.Path)
	}

	return nil
}

// handleDirectory checks if the requested path is a directory and serves the directory listing if it is.
// If the directory has an index.html file, it serves that instead.
// Returns true if the request was handled successfully, false otherwise.
// Returns an error if the request was not handled successfully.
func handleDirectory(w http.ResponseWriter, r *http.Request, reqPath string) (bool, error) {
	fi, _ := os.Stat(reqPath)
	if fi.IsDir() {
		log.Debug("specified path is a directory")
		indexFilePath := filepath.Join(reqPath, "index.html")
		if fi, err := os.Stat(indexFilePath); err != nil {
			// no index.html
			if readmeFilePaths, err := filepath.Glob(filepath.Join(reqPath, "[rR][eE][aA][dD][mM][eE].[mM][dD]")); mdCompile && err == nil && len(readmeFilePaths) != 0 {
				// readme.md exists
				reqPath := readmeFilePaths[0]
				log.Debug("serving readme.md", "reqPath", reqPath)
				fileContent, err := os.ReadFile(reqPath)
				if err != nil {
					return true, errors.Join(fmt.Errorf("failed to read Markdown file %q", reqPath), err)
				}
				if err := md.WriteMarkdown(w, r, r.URL.Path, fileContent); err != nil {
					return true, errors.Join(fmt.Errorf("failed to serve compiled Markdown file %q", reqPath), err)
				}
				return true, nil
			} else {
				err := buildDirectoryListingHTML(w, r, reqPath)
				return err == nil, err
			}
		} else {
			// index.html exists
			if fi.IsDir() {
				// 404: directory called "index.html" has to be explicitly requested
				return false, fmt.Errorf("path %q does not exist", r.URL.Path)
			} else {
				log.Debug("serving index.html")
				// serve index.html
				http.ServeFile(w, r, indexFilePath)
				return true, nil
			}
		}
	}
	return false, nil
}

// handleMarkdown checks if the requested path is a Markdown file and serves the compiled HTML if it is.
// Returns an error if the request was not handled successfully.
func handleMarkdown(w http.ResponseWriter, r *http.Request, reqPath string) error {
	if strings.HasSuffix(r.URL.Path, ".md") {
		fileContent, err := os.ReadFile(reqPath)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to read Markdown file %q", r.URL.Path), err)
		}
		if err := md.WriteMarkdown(w, r, r.URL.Path, fileContent); err != nil {
			return errors.Join(fmt.Errorf("failed to serve compiled Markdown file %q", r.URL.Path), err)
		}
		return nil
	}
	return nil
}

// buildDirectoryListingHTML builds an HTML string for the directory listing of the requested path.
func buildDirectoryListingHTML(w http.ResponseWriter, r *http.Request, reqPath string) error {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("<h1>Directory listing for %s</h1>", r.URL.Path))
	sb.WriteString("<ul>")

	if r.URL.Path != "/" {
		// Add a link to the parent directory
		sb.WriteString(fmt.Sprintf("<a href=%q>../</a><br>", posixpath.Dir(r.URL.Path)))
	}

	dirEntries, _ := os.ReadDir(reqPath)
	for _, dirEntry := range dirEntries {
		display := dirEntry.Name()
		if dirEntry.IsDir() {
			display += "/"
		}
		sb.WriteString(fmt.Sprintf("<a href=%q>%s</a><br>", posixpath.Join("/", r.URL.Path, dirEntry.Name()), display))
	}
	sb.WriteString("</ul>")

	if err := md.WriteDoc(w, r, r.URL.Path, []byte(sb.String())); err != nil {
		return errors.Join(fmt.Errorf("failed to serve directory listing for %q", r.URL.Path), err)
	}
	return nil
}
