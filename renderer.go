package httpserve

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"go/build"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
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

func (s *Server) renderWasmEmbed(pkg string, w http.ResponseWriter, r *http.Request) error {
	log.Printf("building %v...", pkg)

	tf, err := ioutil.TempFile(os.TempDir(), "http-serve.")
	if err != nil {
		return err
	}
	tf.Close()
	defer os.Remove(tf.Name())

	versionCmd := exec.Command("go", "version")
	versionCmd.Stdout = log.Writer()
	versionCmd.Run()

	// BUILDCOMMAND
	errBuf := new(bytes.Buffer)
	cmd := exec.Command("go", "build", "-o", tf.Name(), pkg)
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
	return s.tmpl.ExecuteTemplate(w, "tmpl/wasm_embed.tmpl", map[string]interface{}{
		"wasmexec": template.JS(wasmExec),
		"wasmcode": base64.StdEncoding.EncodeToString(code),
	})
}

// render wasm template or output binary
func (s *Server) renderWasm(pkg string, w http.ResponseWriter, r *http.Request) error {
	if r.URL.Query().Get("f") == "embed" {
		return s.renderWasmEmbed(pkg, w, r)
	}
	if r.URL.Query().Get("f") == "wasm" {
		return s.buildWasm(pkg, w)
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
		"pkg":      pkg,
		"wasmexec": template.JS(wasmExec),
		"wasmfile": pkg + "?f=wasm",
	})
}

// buildWasm builds the wasm file and writes to responseWriter
func (s *Server) buildWasm(pkg string, w http.ResponseWriter) error {
	log.Printf("building %q...", pkg)
	tf, err := ioutil.TempFile(os.TempDir(), "http-serve.")
	if err != nil {
		return err
	}
	if err := tf.Close(); err != nil {
		return err
	}
	defer os.Remove(tf.Name()) // nolint: errcheck

	versionCmd := exec.Command("go", "version")
	versionCmd.Stdout = log.Writer()
	if err := versionCmd.Run(); err != nil {
		return err
	}

	// BUILDCOMMAND
	errBuf := new(bytes.Buffer)
	cmd := exec.Command("go", "build", "-o", tf.Name(), "./"+pkg)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, errBuf)
	if err := cmd.Run(); err != nil {
		return errors.New(errBuf.String())
	}

	f, err := os.Open(tf.Name())
	if err != nil {
		return err
	}
	defer f.Close() // nolint: errcheck
	oi, err := f.Stat()
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", "application/wasm")
	w.Header().Set("Content-length", fmt.Sprint(oi.Size()))
	_, err = io.Copy(w, f)
	return err
}
