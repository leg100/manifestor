package main

import (
	"context"
	"flag"
	"fmt"
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
		Name:          stringPtr(name),
		ID:            int64Ptr(appID),
		PEM:           stringPtr(pem),
		WebhookSecret: stringPtr(webhookSecret),
	}
	// hostname of github server
	githubHostname string
)

func init() {
	flag.IntVar(&port, "port", 0, "Port to listen on; defaults to choosing port dynamically")
	flag.StringVar(&org, "org", "", "Github organization in which to register app. If not specified then the app is registered in your personal account.")
	flag.StringVar(&githubHostname, "hostname", "github.com", "Hostname of github server. Defaults to github.com.")
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

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}
	serverURL := "http://" + listener.Addr().String()

	m, err := newManifestor()
	if err != nil {
		return err
	}
	http.HandleFunc("/", m.submit)
	http.HandleFunc("/complete", m.complete)
	http.HandleFunc("/show", m.show)
	http.HandleFunc("/download", m.download)

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

func stringPtr(s string) *string { return &s }
func int64Ptr(s int64) *int64    { return &s }
