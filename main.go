package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/go-github/v45/github"
	"github.com/pkg/browser"
)

var (
	// listening port - the zero value means port is chosen dynamically.
	port int
	// github organization in which to register app
	org string
	// final app cfg inc. PEM, webhook secret, and App ID
	cfg *github.AppConfig = &github.AppConfig{
		ID:            int64Ptr(appID),
		PEM:           stringPtr(pem),
		WebhookSecret: stringPtr(webhookSecret),
	}
)

func init() {
	flag.IntVar(&port, "port", 0, "Port to listen on; defaults to choosing port dynamically")
	flag.StringVar(&org, "org", "", "Github organization in which to register app. If not specified then the app is registered in your personal account.")
	flag.Parse()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err.Error())
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	catchCtrlC(cancel)

	state, err := generateRandomString(32)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}
	serverURL := "http://" + listener.Addr().String()

	newAppURL := url.URL{
		Scheme:   "https",
		Host:     "github.com",
		Path:     "/settings/apps/new",
		RawQuery: "state=" + state,
	}
	if org != "" {
		newAppURL.Path = "/organizations/" + org + newAppURL.Path
	}

	// The 'submit' page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := renderer.render("submit.tmpl", w, struct {
			State       string
			RedirectURL string
			NameSuffix  string
			NewAppURL   string
		}{
			State:       state,
			RedirectURL: "http://" + r.Host + "/complete",
			NameSuffix:  state[:4],
			NewAppURL:   newAppURL.String(),
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	http.HandleFunc("/complete", func(w http.ResponseWriter, r *http.Request) {
		if state != r.URL.Query().Get("state") {
			http.Error(w, "state paramater mismatch", 500)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "code parameter not found", 500)
			return
		}

		client := github.NewClient(nil)
		var err error
		cfg, _, err = client.Apps.CompleteAppManifest(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		http.Redirect(w, r, "/show", 302)
	})

	http.HandleFunc("/show", func(w http.ResponseWriter, r *http.Request) {
		err := renderer.render("show.tmpl", w, cfg)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	if _, disable := os.LookupEnv("DISABLE_BROWSER"); !disable {
		if err := browser.OpenURL(serverURL); err != nil {
			return err
		}
	}

	go func() {
		http.Serve(listener, nil)
	}()

	<-ctx.Done()
	return ctx.Err()
}

func catchCtrlC(cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	go func() {
		<-signals
		signal.Stop(signals)
		cancel()
	}()
}

func generateRandomString(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func stringPtr(s string) *string { return &s }
func int64Ptr(s int64) *int64    { return &s }
