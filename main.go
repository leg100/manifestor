package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
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
	url := "http://" + listener.Addr().String()
	log.Println("running server at", url)

	// Setup manifest server routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := renderer.render("submit.tmpl", w, struct {
			State       string
			RedirectURL string
			NameSuffix  string
		}{
			State:       state,
			RedirectURL: url + "/complete",
			NameSuffix:  state[:4],
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	http.HandleFunc("/complete", func(w http.ResponseWriter, r *http.Request) {
		if state != r.URL.Query().Get("state") {
			log.Fatal("state paramater mismatch")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			log.Fatal("code parameter not found")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		client := github.NewClient(nil)
		var err error
		cfg, _, err = client.Apps.CompleteAppManifest(r.Context(), code)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully created github app %q. App ID: %d\n", cfg.GetName(), cfg.GetID())

		http.Redirect(w, r, "/show", 302)
	})

	http.HandleFunc("/show", func(w http.ResponseWriter, r *http.Request) {
		err := renderer.render("show.tmpl", w, cfg)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	if _, disable := os.LookupEnv("DISABLE_BROWSER"); !disable {
		if err := browser.OpenURL(url); err != nil {
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
