package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/browser"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err.Error())
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	catchCtrlC(cancel)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return err
	}

	// Setup manifest server routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, nil)
	})

	if err := browser.OpenURL("http://" + listener.Addr().String()); err != nil {
		return err
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
