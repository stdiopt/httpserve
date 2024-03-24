//go:generate go run ./cmd/genversion -out version.go -package httpserve

// Package httpserve serves files, markdown, wasm
package httpserve

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

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

	wasm *WasmHandler
	// Options goes here
}

// New register routes
func New(opt Options) (*Server, error) {
	mux := http.NewServeMux()

	tmpl := template.New("")

	srcFS, err := fs.Sub(assets.FS, "src")
	if err != nil {
		return nil, err
	}

	tmplDir, err := fs.ReadDir(srcFS, "tmpl")
	if err != nil {
		return nil, err
	}
	for _, k := range tmplDir {
		name := k.Name()

		data, err := fs.ReadFile(srcFS, filepath.Join("tmpl", name))
		if err != nil {
			return nil, err
		}

		if _, err := tmpl.New("tmpl/" + name).Parse(string(data)); err != nil {
			return nil, err
		}
	}

	s := &Server{
		Handler:   mux,
		flagMdCSS: opt.FlagMdCSS,
		tmpl:      tmpl,

		wasm: &WasmHandler{tmpl},
	}

	mux.HandleFunc("/.httpServe/_reload", s.watcher)
	mux.HandleFunc("/.httpServe/d2", s.renderD2)
	mux.Handle("/.httpServe/", http.StripPrefix("/.httpServe", http.FileServer(http.FS(srcFS))))
	mux.HandleFunc("/", s.renderer)

	return s, nil
}

// The default route
func (s Server) renderer(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]

	if path == "" {
		path = "." // Cur dir
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path = filepath.Clean(path)
	ext := filepath.Ext(path)

	fstat, err := os.Stat(path)
	if err != nil {
		writeStatus(w, http.StatusNotFound)
		return
	}
	raw := r.URL.Query().Get("raw")

	switch {
	case fstat.IsDir() && raw == "1":
		if err := s.renderFolder(path, w, r); err != nil {
			writeStatus(w, http.StatusInternalServerError, err)
		}
	case fstat.IsDir():
		// Check for index file
		indexFile := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexFile); err == nil {
			http.ServeFile(w, r, indexFile)
			break
		}
		// Check for main.go file
		mainGo := filepath.Join(path, "main.go")
		if _, err := os.Stat(mainGo); err == nil {
			if err := s.wasm.render(path, w, r); err != nil {
				writeStatus(w, http.StatusInternalServerError, err)
			}
			break
		}
		if err := s.renderFolder(path, w, r); err != nil {
			writeStatus(w, http.StatusInternalServerError, err)
		}
	case raw == "1":
		http.ServeFile(w, r, path)
	case ext == ".md":
		if err := s.renderFileMarkDown(path, w, r); err != nil {
			writeStatus(w, http.StatusInternalServerError, err)
		}
	case ext == ".dot" && r.URL.Query().Get("f") == "png":
		if err := s.renderFileDotPng(path, w, r); err != nil {
			writeStatus(w, http.StatusInternalServerError, err)
		}
	case ext == ".d2" && r.URL.Query().Get("f") == "png":
		if err := s.renderFileD2(path, w, r); err != nil {
			writeStatus(w, http.StatusInternalServerError, err)
		}
	default:
		http.ServeFile(w, r, path)
	}
}

func writeStatus(w http.ResponseWriter, code int, extras ...interface{}) {
	w.WriteHeader(code)
	extra := fmt.Sprint(extras...)
	fmt.Fprint(w, http.StatusText(code), extra)
}
