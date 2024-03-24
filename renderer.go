package httpserve

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	blackfriday "github.com/russross/blackfriday/v2"
)

func (s *Server) renderFileMarkDown(p string, w http.ResponseWriter, r *http.Request) error {
	fileData, err := os.ReadFile(p)
	if err != nil {
		return err
	}

	opt := blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.HeadingIDs | blackfriday.AutoHeadingIDs)

	outputHTML := blackfriday.Run(fileData, opt)

	w.Header().Set("Content-type", "text/html")
	err = s.tmpl.ExecuteTemplate(w, "tmpl/markdown.tmpl", map[string]interface{}{
		"rand":       rand.Int(),
		"css":        s.flagMdCSS,
		"path":       p,
		"outputHTML": template.HTML(string(outputHTML)),
	})
	return err
}

func (s *Server) renderFolder(p string, w http.ResponseWriter, r *http.Request) error {
	res, err := os.ReadDir(p)
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", "text/html")
	err = s.tmpl.ExecuteTemplate(w, "tmpl/folder.tmpl", map[string]interface{}{
		"path":    p,
		"content": res,
	})
	return err
}

// Execute command `dot`
func (s *Server) renderFileDotPng(p string, w http.ResponseWriter, r *http.Request) error {
	absPath, err := filepath.Abs(p)
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", "image/png")
	cmd := exec.Command("dot", "-Tpng", absPath)
	cmd.Stdout = w
	return cmd.Run()
}

func (s *Server) renderFileD2(p string, w http.ResponseWriter, r *http.Request) error {
	log.Println("Rendering D2", p)
	absPath, err := filepath.Abs(p)
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", "image/svg+xml")
	f, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("d2", "-")
	cmd.Stdout = w
	cmd.Stdin = f
	return cmd.Run()
}

// Receive d2 file and render
func (s *Server) renderD2(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-type", "image/svg+xml")
	cmd := exec.Command("d2", "--pad", "0", "-")
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		writeStatus(w, http.StatusInternalServerError, err.Error())
	}
}
