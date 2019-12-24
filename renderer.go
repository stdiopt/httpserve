package httpserve

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"go/build"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/exp/rand"
	blackfriday "gopkg.in/russross/blackfriday.v2"
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

func (s *Server) renderWasm(p string, w http.ResponseWriter, r *http.Request) error {
	log.Printf("building %v...", p)

	tf, err := ioutil.TempFile(os.TempDir(), "http-serve.")
	if err != nil {
		return err
	}
	tf.Close()

	defer os.Remove(tf.Name())

	// BUILDCOMMAND
	errBuf := new(bytes.Buffer)
	cmd := exec.Command("go", "build", "-o", tf.Name(), "./"+p)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = errBuf
	if err := cmd.Run(); err != nil {
		fmt.Fprint(w, errBuf.String())
		return err
	}

	// wasm code read
	code, err := ioutil.ReadFile(tf.Name())
	if err != nil {
		log.Println("err:", err)
		return err
	}

	goroot := build.Default.GOROOT
	wasmExecName := filepath.Join(goroot, "misc/wasm/wasm_exec.js")

	// Read wasm_exec from system dist
	wasmExec, err := ioutil.ReadFile(wasmExecName)
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", "text/html")
	return s.tmpl.ExecuteTemplate(w, "tmpl/wasm.tmpl", map[string]interface{}{
		"wasmexec": template.JS(wasmExec),
		"wasmcode": base64.StdEncoding.EncodeToString(code),
	})
}
