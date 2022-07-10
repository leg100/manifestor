package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/v45/github"
)

type manifestor struct {
	state string
	cfg   *github.AppConfig
}

func newManifestor() (*manifestor, error) {
	state, err := generateRandomString(32)
	if err != nil {
		return nil, err
	}

	return &manifestor{
		state: state,
	}, nil
}

func (m *manifestor) State() string      { return m.state }
func (m *manifestor) NameSuffix() string { return m.state[:4] }

func (m *manifestor) submit(w http.ResponseWriter, r *http.Request) {
	newAppURL := url.URL{
		Scheme:   "https",
		Host:     githubHostname,
		Path:     "/settings/apps/new",
		RawQuery: "state=" + m.state,
	}
	if org != "" {
		newAppURL.Path = "/organizations/" + org + newAppURL.Path
	}

	manifest := newManifest(m.NameSuffix(), "http://"+r.Host+"/complete", "http://doesnotexist.com")
	marshalled, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	err = renderer.render("submit.tmpl", w, struct {
		NewAppURL string
		Manifest  string
	}{
		NewAppURL: newAppURL.String(),
		Manifest:  string(marshalled),
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (m *manifestor) complete(w http.ResponseWriter, r *http.Request) {
	if m.state != r.URL.Query().Get("state") {
		http.Error(w, "state parameter mismatch", 500)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code parameter not found", 500)
		return
	}

	client := github.NewClient(nil)
	cfg, _, err := client.Apps.CompleteAppManifest(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	m.cfg = cfg

	http.Redirect(w, r, "/show", 302)
}

func (m *manifestor) show(w http.ResponseWriter, r *http.Request) {
	if err := renderer.render("show.tmpl", w, m.cfg); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (m *manifestor) download(w http.ResponseWriter, r *http.Request) {
	fname := m.cfg.GetName() + ".pem"
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", fname))
	w.Write([]byte(m.cfg.GetPEM()))
}

func generateRandomString(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
