package httpserve

import (
	"bytes"
	"encoding/base64"
	"errors"
	"go/build"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// TODO: split this into into a new handler:

type WasmHandler struct {
	tmpl *template.Template
}

func (h *WasmHandler) tmpFile() (*os.File, func(), error) {
	tf, err := os.CreateTemp("", "http-serve.")
	if err != nil {
		return nil, nil, err
	}

	clean := func() {
		tf.Close()
		defer os.Remove(tf.Name())
	}

	return tf, clean, nil
}

// render wasm template or output binary
func (h *WasmHandler) render(pkg string, w http.ResponseWriter, r *http.Request) error {
	if r.URL.Query().Get("f") == "embed" {
		return h.renderEmbed(pkg, w, r)
	}
	if r.URL.Query().Get("f") == "wasm" {
		f, err := h.build(pkg)
		if err != nil {
			return err
		}
		defer f.Close()
		http.ServeFile(w, r, f.Name())
		return nil
	}

	goroot := build.Default.GOROOT
	wasmExecName := filepath.Join(goroot, "misc/wasm/wasm_exec.js")

	// Read wasm_exec from system dist
	wasmExec, err := os.ReadFile(wasmExecName)
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", "text/html")
	return h.tmpl.ExecuteTemplate(w, "tmpl/wasm.tmpl", map[string]interface{}{
		"pkg":      pkg,
		"wasmexec": template.JS(wasmExec),
		"wasmfile": pkg + "?f=wasm",
	})
}

func (h *WasmHandler) renderEmbed(pkg string, w http.ResponseWriter, r *http.Request) error {
	log.Printf("building %v...", pkg)

	f, err := h.build(pkg)
	if err != nil {
		return err
	}
	defer f.Close()

	code, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	goroot := build.Default.GOROOT
	wasmExecName := filepath.Join(goroot, "misc/wasm/wasm_exec.js")

	// Read wasm_exec from system dist
	wasmExec, err := os.ReadFile(wasmExecName)
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", "text/html")
	return h.tmpl.ExecuteTemplate(w, "tmpl/wasm_embed.tmpl", map[string]interface{}{
		"wasmexec": template.JS(wasmExec),
		"wasmcode": base64.StdEncoding.EncodeToString(code),
	})
}

// buildWasm builds the wasm file and returns the temporary result file
// file is deleted on close
func (h *WasmHandler) build(pkg string) (*tmpFile, error) {
	log.Printf("building %q...", pkg)

	tf, err := os.CreateTemp(os.TempDir(), "http-serve.")
	if err != nil {
		return nil, err
	}
	if err := tf.Close(); err != nil {
		return nil, err
	}

	versionCmd := exec.Command("go", "version")
	versionCmd.Stdout = log.Writer()
	if err := versionCmd.Run(); err != nil {
		return nil, err
	}

	// BUILDCOMMAND
	errBuf := new(bytes.Buffer)
	cmd := exec.Command("go", "build", "-o", tf.Name(), "./"+pkg)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, errBuf)
	if err := cmd.Run(); err != nil {
		return nil, errors.New(errBuf.String())
	}

	f, err := os.Open(tf.Name())
	if err != nil {
		return nil, err
	}
	return &tmpFile{f}, nil
}

// Helper
type tmpFile struct {
	*os.File
}

// Close will remove the file regardles of the closing error
func (t *tmpFile) Close() error {
	defer os.Remove(t.Name()) // nolint: errcheck
	return t.File.Close()
}
