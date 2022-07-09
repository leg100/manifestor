package main

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"os"
	"text/template"
)

// This file handles serving static files (css etc) and parsing and rendering
// templates. A development mode is supported to permit a developer to make
// edits to static files and templates and see the resulting changes in a
// browser without re-compiling the go binary.

var (
	// static files embedded within the go binary
	//
	//go:embed static
	embedded embed.FS
	// same files but on the local disk
	localDisk = os.DirFS(".")
	// filesystem containing static files; defaults to serving up embedded
	fsys fs.FS = embedded
	// template renderer
	renderer *templateRenderer
)

func init() {
	// Toggling dev mode instead serves and renders files from disk
	if _, devMode := os.LookupEnv("DEV_MODE"); devMode {
		fsys = localDisk
	}
	// serve static files (css, js, etc) from fs
	http.Handle("/static/", http.FileServer(http.FS(fsys)))
	// render templates from fs
	renderer = &templateRenderer{fsys}
}

type templateRenderer struct {
	fs.FS
}

func (r *templateRenderer) render(name string, w io.Writer, data any) error {
	tpl, err := template.ParseFS(fsys, "static/"+name, "static/layout.tmpl")
	if err != nil {
		return err
	}
	return tpl.ExecuteTemplate(w, "layout", data)
}
