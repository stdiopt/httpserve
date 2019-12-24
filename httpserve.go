//go:generate go run ./cmd/genversion -out version.go -package httpserve
//go:generate go run github.com/gohxs/folder2go -nobackup assets-src assets

package httpserve

import (
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/stdiopt/httpserve/assets"
)

// Options for httpserve
type Options struct {
	FlagMdCSS string
}

// Server thing
type Server struct {
	http.Handler
	flagMdCSS string
	tmpl      *template.Template
	// Options goes here
}

// New register routes
func New(opt Options) *Server {
	mux := http.NewServeMux()

	tmpl := template.New("")
	for k := range assets.Data {
		if !strings.HasPrefix(k, "tmpl/") {
			continue
		}
		_, err := tmpl.New(k).Parse(string(assets.Data[k]))
		if err != nil {
			log.Fatal("Internal error, loading templates")
		}
	}

	s := &Server{
		Handler:   mux,
		flagMdCSS: opt.FlagMdCSS,
		tmpl:      tmpl,
	}

	mux.HandleFunc("/.httpServe/_reload", s.watcher)
	mux.Handle("/.httpServe/", http.StripPrefix("/.httpServe", http.HandlerFunc(s.binData)))
	mux.HandleFunc("/", s.files)

	return s

}

// The default route
func (s Server) files(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]

	if path == "" {
		path = "." // Cur dir
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if strings.Contains(path, "..") { // ServeFile will normalize path
		http.ServeFile(w, r, path)
	}

	fstat, err := os.Stat(path)
	if err != nil {
		writeStatus(w, http.StatusNotFound)
		return
	}
	raw := r.URL.Query().Get("raw")

	// Rules to select file renderers

	// It is a dir
	// Handle dir in another method
	if fstat.IsDir() {
		if raw != "1" {
			// Check for index file
			indexFile := filepath.Join(path, "index.html")
			if _, err := os.Stat(indexFile); err == nil {
				http.ServeFile(w, r, indexFile)
				return
			}
			// Check for main.go file
			mainGo := filepath.Join(path, "main.go")
			if _, err := os.Stat(mainGo); err == nil {
				if err := s.renderWasm(path, w, r); err != nil {
					writeStatus(w, http.StatusInternalServerError, err)
				}
				return
			}
		}
		if err := s.renderFolder(path, w, r); err != nil {
			writeStatus(w, http.StatusInternalServerError, err)
		}
		return
	}

	if raw == "1" {
		http.ServeFile(w, r, path)
	}

	if filepath.Ext(path) == ".md" {
		if err := s.renderMarkDown(path, w, r); err != nil {
			writeStatus(w, http.StatusInternalServerError, err)
		}
		return
	}
	if filepath.Ext(path) == ".dot" && r.URL.Query().Get("f") == "png" {
		if err := s.renderDotPng(path, w, r); err != nil {
			writeStatus(w, http.StatusInternalServerError, err)
		}
		return
	}
	// default
	http.ServeFile(w, r, path)
}

// binData handler
func (s Server) binData(w http.ResponseWriter, r *http.Request) {
	urlPath := strings.TrimPrefix(r.URL.String(), "/")
	if urlPath == "" {
		urlPath = "index.html"
	}
	data, ok := assets.Data[urlPath]
	if !ok {

		writeStatus(w, http.StatusNotFound, "Not found")
		return
	}
	w.Header().Set("Content-type", mime.TypeByExtension(filepath.Ext(urlPath)))
	w.Write(data)
}

func writeStatus(w http.ResponseWriter, code int, extras ...interface{}) {
	w.WriteHeader(code)
	extra := fmt.Sprint(extras...)
	fmt.Fprint(w, http.StatusText(code), extra)
}
