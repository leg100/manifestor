package main

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"text/template"
)

var (
	// Files embedded within the go binary
	//
	//go:embed static
	embedded embed.FS

	// The same files but on the local disk
	localDisk = os.DirFS(".")

	// the parsed index.html template
	tpl *template.Template
)

func init() {
	// filesystem containing static files; defaults to serving up embedded
	var fsys fs.FS = embedded

	// Toggling dev mode enables serving files up from local disk instead.
	if _, devMode := os.LookupEnv("DEV_MODE"); devMode {
		fsys = localDisk
	}

	http.Handle("/static", http.FileServer(http.FS(fsys)))
	tpl = template.Must(template.ParseFS(fsys, "static/*.tmpl"))
}
