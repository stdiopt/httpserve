package httpserve

import (
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"

	blackfriday "github.com/russross/blackfriday/v2"
)

func (s *Server) renderMarkDown(p string, w http.ResponseWriter, r *http.Request) error {
	fileData, err := ioutil.ReadFile(p)
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
	res, err := ioutil.ReadDir(p)
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
func (s *Server) renderDotPng(p string, w http.ResponseWriter, r *http.Request) error {
	absPath, err := filepath.Abs(p)
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", "image/png")
	cmd := exec.Command("dot", "-Tpng", absPath)
	cmd.Stdout = w
	return cmd.Run()
}
