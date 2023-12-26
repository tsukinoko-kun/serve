package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Frank-Mayer/serve/internal/md"
	"github.com/Frank-Mayer/serve/internal/utils"
	"github.com/Frank-Mayer/serve/internal/version"
	"github.com/charmbracelet/log"
	"go.chromium.org/goma/server/command/descriptor/posixpath"
)

var (
	dirPath    = flag.String("dir", ".", "Path to the directory to serve")
	absPath    string
	mdCompile  = flag.Bool("md", false, "Compile Markdown files to HTML")
	port       = flag.Int("port", 8080, "Port to listen on")
	verb       = flag.Bool("verbose", false, "Verbose output")
	vers       = flag.Bool("version", false, "Print version")
	fileServer = http.FileServer(http.Dir(*dirPath))
)

func main() {
	flag.Parse()

	if *vers {
		fmt.Println(version.Version)
		return
	}

	if *verb {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Check if the directory exists
	fi, err := os.Stat(*dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("Directory does not exist", "path", *dirPath)
		}
		log.Fatal("Error checking if directory exists", "path", *dirPath, "error", err)
	}

	// Check if the path is a directory
	if !fi.IsDir() {
		log.Fatalf("%s is not a directory", *dirPath)
	}

	// Abs path
	absPath, err = filepath.Abs(*dirPath)
	if err != nil {
		log.Fatal("Error getting absolute path", "path", *dirPath, "error", err)
	}

	// Middleware to compile Markdown files if the -md flag is set
	http.Handle("/", http.HandlerFunc(handler))

	// Start the HTTP server
	log.Info("Serving", "path", absPath, "port", *port, "md", *mdCompile)
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	reqPath := filepath.Join(absPath, r.URL.Path)

	log.Debug("Request", "path", reqPath)

	// Check if the file exists
	fi, err := os.Stat(reqPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("File %s does not exist", r.URL.Path), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Error checking if file %s exists: %v", r.URL.Path, err), http.StatusInternalServerError)
		return
	}

	// Check if the file is in the directory
	if !utils.IsIn(reqPath, *dirPath) {
		http.Error(w, fmt.Sprintf("File %s does not exist", r.URL.Path), http.StatusNotFound)
		return
	}

	// Check if the file is a directory
	if fi.IsDir() {
		// Check if the directory has an index.html file
		indexFilePath := filepath.Join(reqPath, "index.html")
		_, err := os.Stat(indexFilePath)
		if err != nil {
			// If the directory does not have an index.html file, serve the directory listing
			sb := strings.Builder{}
			dirEntries, err := os.ReadDir(reqPath)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reading directory: %v", err), http.StatusInternalServerError)
				return
			}
			sb.WriteString(fmt.Sprintf("<h1>Directory listing for %s</h1>", r.URL.Path))
			sb.WriteString("<ul>")
			for _, dirEntry := range dirEntries {
				sb.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a><br>", posixpath.Join("/", r.URL.Path, dirEntry.Name()), dirEntry.Name()))
			}
			sb.WriteString("</ul>")
			w.Header().Set("Content-Type", "text/html")
			_, err = w.Write(md.Compile([]byte(sb.String())))
			if err != nil {
				http.Error(w, fmt.Sprintf("Error serving directory listing: %v", err), http.StatusInternalServerError)
				return
			}
			return
		}
		reqPath = indexFilePath
	}

	// Check if the file is a Markdown file
	if *mdCompile && strings.HasSuffix(r.URL.Path, ".md") {
		fileContent, err := os.ReadFile(reqPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading Markdown file: %v", err), http.StatusInternalServerError)
			return
		}
		htmlContent := md.Compile(fileContent)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write(htmlContent)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error serving HTML: %v", err), http.StatusInternalServerError)
			return
		}
		return
	}
	fileServer.ServeHTTP(w, r)
}
